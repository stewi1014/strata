package sync

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
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

// prepend adds one element to the start of the slice
func prepend[E any, S []E](s *S, elem E) {
	if len(*s) == 0 {
		*s = []E{elem}
		return
	}

	if cap(*s) > len(*s) {
		*s = (*s)[:len(*s)+1]
		copy((*s)[1:], *s)
		(*s)[0] = elem
		return
	}

	new := make(S, len(*s)+1, 2*cap(*s))
	copy(new[1:], *s)
	new[0] = elem
	*s = new
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
	parent *Sync
	key    any

	// protected by mutex
	mutex   sync.RWMutex
	child   map[any]*Sync
	clients []*client
}

type client struct {
	enc  *gob.Encoder
	dec  *gob.Decoder
	ctx  context.Context
	done func(error)
}

func (s *Sync) isZero() bool {
	return s == nil || (s.vptr.Load() == nil &&
		len(s.child) == 0 &&
		len(s.clients) == 0)
}

// returns the root node, along with the key of the current node
// relative to the root node.
func (s *Sync) root() (root *Sync, key []any) {
	for s.parent != nil {
		prepend(&key, s.key)
		s = s.parent
	}

	return s, key
}

func (s *Sync) sendAll(client *client, key ...any) {
	ptr := s.vptr.Load()
	if ptr != nil {
		v, t := (*ptr).GetAt()
		client.send(message{
			V: v,
			T: t,
			K: key,
		})
	}

	s.mutex.RLock()
	// this approach is taken to
	// avoid doing the send while holding mutex
	// or writing to the client with a great number of routines
	children := make([]*Sync, len(s.child))
	i := 0
	for _, v := range s.child {
		children[i] = v
		i++
	}
	s.mutex.RUnlock()

	for i := range children {
		children[i].sendAll(client, append(key, children[i].key)...)
	}
}

// get returns the node at the given key,
// if it exists.
func (s *Sync) get(key ...any) *Sync {
	for i := 0; i < len(key) && s != nil; i++ {
		s.mutex.RLock()
		next := s.child[key[i]]
		s.mutex.RUnlock()
		s = next
	}

	return s
}

// Delete removes the value at the given key,
// internally cleaning up unused nodes.
//
// It's the inverse of Register.
func (s *Sync) Delete(key ...any) {
	// navigate to the end,
	// removing the value if it exists
	for i := 0; ; i++ {
		if i >= len(key) {
			// at the end
			s.vptr.Store(nil)
			break
		}

		s.mutex.RLock()
		next := s.child[key[i]]
		s.mutex.RUnlock()

		if next == nil {
			break
		}

		s = next
	}

	s.mutex.Lock()
	for s.isZero() && s.parent != nil {
		s.parent.mutex.Lock()
		delete(s.parent.child, s.key)
		s.mutex.Unlock()
		s = s.parent
	}
	s.mutex.Unlock()
}

// Branch returns the node at the given key,
// creating nodes as needed to reach it.
func (s *Sync) Branch(key ...any) *Sync {
	for i := 0; i < len(key); i++ {
		s.mutex.RLock()
		next := s.child[key[i]]
		s.mutex.RUnlock()

		if next == nil {
			s.mutex.Lock()
			next = s.child[key[i]]

			// check again,
			// as another thread may have created it during the
			// read -> write lock transition.
			if next == nil {
				if s.child == nil {
					s.child = make(map[any]*Sync)
				}

				next = new(Sync)
				next.parent = s
				next.key = key[i]
				s.child[key[i]] = next
			}

			s.mutex.Unlock()
		}

		s = next
		key = key[1:]
	}

	return s
}

func (s *Sync) Sync(ctx context.Context, rw io.ReadWriter) error {
	c := s.newClient(ctx, gob.NewEncoder(rw), gob.NewDecoder(rw))

	go s.listen(c)
	go s.sendAll(c)

	<-c.ctx.Done()
	return context.Cause(c.ctx)
}

func (s *Sync) SyncTo(ctx context.Context, w io.Writer) error {
	c := s.newClient(ctx, gob.NewEncoder(w), nil)

	go s.sendAll(c)

	<-c.ctx.Done()
	return context.Cause(c.ctx)
}

func (s *Sync) SyncFrom(ctx context.Context, r io.Reader) error {
	c := s.newClient(ctx, nil, gob.NewDecoder(r))

	go s.listen(c)

	<-c.ctx.Done()
	return context.Cause(c.ctx)
}

func (s *Sync) newClient(ctx context.Context, enc *gob.Encoder, dec *gob.Decoder) *client {
	ctx, done := context.WithCancelCause(ctx)

	c := &client{
		enc:  enc,
		dec:  dec,
		ctx:  ctx,
		done: done,
	}

	s.mutex.Lock()
	s.clients = append(s.clients, c)
	s.mutex.Unlock()

	context.AfterFunc(ctx, func() {
		s.mutex.Lock()
		del(&s.clients, c)
		s.mutex.Unlock()
	})

	return c
}

func (s *Sync) listen(client *client) {
	for {
		var msg message
		err := client.dec.Decode(&msg)
		if err != nil {
			client.done(err)
			return
		}

		if err := client.ctx.Err(); err != nil {
			return
		}

		node := s.get(msg.K...)
		if node == nil {
			// don't have the value
			continue
		}

		v := node.vptr.Load()
		if v == nil {
			// don't have the value
			continue
		}

		value, t := (*v).GetAt()
		if t.After(msg.T) {
			// already have a newer value
			// reply with our newer value
			//
			// this also allows new clients to
			// query current values by sending a
			// zero time.

			msg.T = t
			msg.V = value
			client.send(msg)
			continue
		}

		if reflect.TypeOf(value) != reflect.TypeOf(msg.V) {
			client.done(fmt.Errorf(
				"bad type for %v received: got %T but have %T",
				msg.K,
				msg.V,
				value,
			))

			return
		}

		// propagate the change
		(*v).SetAt(msg.V, msg.T)
	}
}

func (c *client) send(msg message) {
	if c.enc == nil {
		return
	}

	if c.ctx.Err() != nil {
		return
	}

	err := c.enc.Encode(&msg)
	if err != nil {
		c.done(err)
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

	s = s.Branch(v.Type().String())

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		if impl := f.Addr(); impl.Type().Implements(valuelikeType) {
			name := fmt.Sprintf("[%v]%v", i, v.Type().Field(i).Name)
			s.Branch(name).register(impl.Interface().(Valuelike))
			continue
		}

		if impl := f; impl.Type().Implements(valuelikeType) {
			name := fmt.Sprintf("[%v]%v", i, v.Type().Field(i).Name)
			s.Branch(name).register(impl.Interface().(Valuelike))
			continue
		}
	}
}

func (s *Sync) Register(value Valuelike, key ...interface{}) {
	s.Branch(key...).register(value)
}

func (s *Sync) register(value Valuelike) {
	if !s.vptr.CompareAndSwap(nil, &value) {
		panic(fmt.Errorf("%v is already registered", value))
	}

	value.OnChange(s.onChange)
	s.onChange(value, time.Time{})
}

func (s *Sync) onChange(value any, when time.Time) {
	s, key := s.root()

	for i := range key {
		s.mutex.RLock()
		for c := range s.clients {
			go s.clients[c].send(message{
				T: when,
				K: key[i:],
				V: value,
			})
		}

		next := s.child[key[i]]
		s.mutex.RUnlock()

		s = next
	}
}
