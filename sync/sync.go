package sync

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// del deletes one element from a slice,
// returning its index.
func del[E comparable, S []E](s *S, elem E) int {
	for i, e := range *s {
		if e == elem {
			copy((*s)[i:], (*s)[i+1:])
			*s = (*s)[:len(*s)-1]
			return i
		}
	}

	return -1
}

// message is the type given to encoding/gob
// for communication between Sync instances.
type message struct {
	T time.Time
	K []any
	V any
}

// Sync represents a tree-like structure that synchronises across instances using io.Writer/io.Reader.
// The zero state is ready for use.
//
// The general motivation behind the tree structure is that it provides the most generalized
// identification method for values while remaining performant.
// e.g. a node might use a struct type name as its key, with
// each child node representing a field within that struct, identified by name (see RegisterStruct).
//
// Clients can be segregated to specific branches too. Sync.Branch("myBranch").Sync(conn) for example
// will synchronise only the "myBranch" node and its children with conn, with "myBranch" being
// the root node that the other Sync instance (behind conn) will see.
type Sync struct {
	vptr atomic.Pointer[Valuelike]

	// read only
	// parent remains even after deletion
	// see touch
	parent *Sync
	key    any

	// protected by mutex
	mutex   sync.RWMutex
	child   map[any]*Sync
	clients []*client
}

func (s *Sync) isZero() bool {
	return s == nil || (s.vptr.Load() == nil &&
		len(s.child) == 0 &&
		len(s.clients) == 0)
}

func (s *Sync) walk(f func(s *Sync, key []any), prefix ...any) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	f(s, prefix)

	for k, v := range s.child {
		v.walk(f, append(prefix, k)...)
	}
}

// get returns the value at the given key,
// returning nil if it doesn't exist.
func (s *Sync) get(key ...any) Valuelike {
	s.mutex.RLock()
	for i := range key {
		next := s.child[key[i]]
		if next == nil {
			s.mutex.RUnlock()
			return nil
		}

		next.mutex.RLock()
		s.mutex.RUnlock()
		s = next
	}
	s.mutex.RUnlock()

	ptr := s.vptr.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}

// with allows modifications to a node at a given key.
//
// The write mutex is held, allowing modifications to the node,
// and the key of the node relative to the root is given.
//
// The node and any connectiosn are created if needed.
func (s *Sync) with(key []any, f func(s *Sync)) {
	// we could be on a branch removed from the tree,
	// and someone, somewhere, kept a reference
	// and now wants to modify a part of it.
	//
	// that being said, if this is part of a disconnected branch,
	// then it should be a single chain of zero-values.
	//
	// where does that leave us?
	// we can go up, populate what we need, but nothing that matters can access it anymore.
	// running without error is disingenuous,
	// and the caller wouldn't necessarily have a reasonable way to mitigate this.
	//
	// instead I think this motivates parent and key never being touched after creation,
	// and reconnecting this branch to the tree.
	// Even if it's empty, the caller still has the reference and wants to make calls.
	//
	// no way to know where the break is
	// or if there's two of them!
	// so, get to root
	for s.parent != nil {
		key = append([]any{s.key}, key...) // don't modify key's original backing array
		s = s.parent
	}

	s.mutex.Lock()
	for i := range key {
		next := s.child[key[i]]
		if next == nil {
			if s.child == nil {
				s.child = make(map[any]*Sync, 1)
			}

			next = &Sync{
				parent: s,
				key:    key[i],
			}
			s.child[key[i]] = next
		}

		s = next
		s.mutex.Lock()
		s.parent.mutex.Unlock()
	}

	defer s.mutex.Unlock()
	f(s)
}

func newClient(ctx context.Context) *client {
	ctx, done := context.WithCancelCause(ctx)
	return &client{
		ctx:  ctx,
		done: done,
		send: make(chan *message),
	}
}

type client struct {
	ctx  context.Context
	done func(error)
	send chan *message
}

func (s *Sync) Sync(conn net.Conn) error {
	c := newClient(context.Background())
	go s.syncFrom(c, conn)
	go s.syncTo(c, conn)

	<-c.ctx.Done()
	return context.Cause(c.ctx)
}

func (s *Sync) SyncFrom(r io.Reader) error {
	c := newClient(context.Background())
	return s.syncFrom(c, r)
}

func (s *Sync) syncFrom(c *client, r io.Reader) error {
	dec := gob.NewDecoder(r)
	var msg message

	for {
		err := dec.Decode(&msg)
		if err != nil {
			c.done(err)
			return err
		}

		err = s.recvMessage(c, msg)
		if err != nil {
			c.done(err)
			return err
		}
	}
}

func (s *Sync) recvMessage(c *client, msg message) error {
	v := s.get(msg.K...)
	if v == nil {
		return nil
	}

	value, t := v.GetAt()
	if t.After(msg.T) {
		// already have a newer value
		// reply with our newer value
		//
		// this also allows new clients to
		// query current values by sending a
		// zero time.

		msg.T = t
		msg.V = value
		s.sendMessage(c, msg)
		return nil
	}

	if reflect.TypeOf(value) != reflect.TypeOf(msg.V) {
		return fmt.Errorf(
			"bad type for %v received: got %T but have %T",
			msg.K,
			msg.V,
			value,
		)
	}

	v.SetAt(msg.V, msg.T)
	return nil
}

func (s *Sync) SyncTo(w io.Writer) error {
	c := newClient(context.Background())
	return s.syncTo(c, w)
}

func (s *Sync) syncTo(c *client, w io.Writer) error {
	s.mutex.Lock()
	s.clients = append(s.clients, c)
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		del(&s.clients, c)
		s.mutex.Unlock()

		c.done(fmt.Errorf("unknown cause"))
	}()

	go s.walk(func(s *Sync, key []any) {
		if ptr := s.vptr.Load(); ptr != nil {
			v, t := (*ptr).GetAt()
			c.send <- &message{
				T: t,
				K: key,
				V: v,
			}
		}
	})

	enc := gob.NewEncoder(w)
	for {
		err := enc.Encode(<-c.send)
		if err != nil {
			c.done(err)
			return err
		}
	}
}

func (s *Sync) sendMessage(c *client, msg message) {
	if c.send == nil {
		return
	}

	select {
	case <-c.ctx.Done():
	case c.send <- &msg:
	}
}

func (s *Sync) RegisterStruct(structPtr any) {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%v is not a pointer to struct", v.Type()))
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("%v is not a struct", v.Type()))
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		if impl := f.Addr(); impl.Type().Implements(valuelikeType) {
			key := []any{
				v.Type().String(),
				fmt.Sprintf("[%v]%v", i, v.Type().Field(i).Name),
			}

			s.Register(impl.Interface().(Valuelike), key...)

			continue
		}

		if impl := f; impl.Type().Implements(valuelikeType) {
			key := []any{
				v.Type().String(),
				fmt.Sprintf("[%v]%v", i, v.Type().Field(i).Name),
			}

			s.Register(impl.Interface().(Valuelike), key...)

			continue
		}
	}
}

func (s *Sync) Register(value Valuelike, key ...interface{}) {
	s.with(key, func(s *Sync) {
		if !s.vptr.CompareAndSwap(nil, &value) {
			panic(fmt.Errorf("%v is already registered", value))
		}

		value.OnChange(s.onChange)
	})

	v, t := value.GetAt()
	s.onChange(v, t)
}

func (s *Sync) onChange(value any, when time.Time) {
	var key []any

	s.mutex.RLock()
	for {
		for c := range s.clients {
			go s.sendMessage(
				s.clients[c],
				message{
					T: when,
					K: key,
					V: value,
				},
			)
		}

		if s.parent == nil {
			s.mutex.RUnlock()
			return
		}

		key = append([]any{s.key}, key...)

		s.parent.mutex.RLock()
		s.mutex.RUnlock()
		s = s.parent
	}
}
