// sync is an attempt at some kind of key-value style thread-safe distributed state machine.
//
// All values must be gob encodable (register your types).
package sync

import (
	"context"
	"encoding/gob"
	"net"
	"sync"
	"time"
)

// message holds extra data that is needed when propagating value changes.
type message struct {
	t     time.Time
	keys  []interface{}
	value interface{}
}

// Syncrhoniser needs no initialisation
//
// Except it kinda does - let the first call to Sync do it.
type Syncrhoniser struct {
	mutex sync.RWMutex

	send           chan message
	numConnections uint
}

// Sync sychronises changes with the given ReadWriter.
//
// It blocks until error.
func (s *Syncrhoniser) Sync(conn net.Conn) error {
	s.mutex.Lock()
	if s.numConnections == 0 {
		s.send = make(chan message)
	}
	s.numConnections++
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		if s.numConnections == 1 {
			close(s.send)
		}
		s.numConnections--
		s.mutex.Unlock()
	}()

	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	ctx, cancel := context.WithCancelCause(context.Background())

	// read
	go func() {
		message := new(message)

	}()

	// write
	go func() {

	}()

	<-ctx.Done()
	return context.Cause(ctx)
}

func (s *Syncrhoniser) read(dec *gob.Decoder) {

}

func (s *Syncrhoniser) write(enc *gob.Encoder) {

}

func (s *Syncrhoniser) Register(value Value[any], key ...interface{}) {

}

type Value[T any] struct {
	lastChange time.Time
	value      T
}

func (v *Value[T]) Get() T {
	return v.value
}

func (v *Value[T]) Set(value T) {

}

// OnChange registers a function that is called when the value changes.
//
// It returns a function that removes this registration.
func (v *Value[T]) OnChange(onChange func(T)) (cancel func()) {
	return nil
}
