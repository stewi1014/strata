package strata

import (
	"encoding/gob"
	"io"
	"sync"
	"time"
)

type Syncrhoniser struct {
	t *Tree[syncValue]

	initOnce  sync.Once
	send      chan<- *syncMsg
	newSender func() (<-chan *syncMsg, func())
}

type syncMsg struct {
	T     time.Time
	Key   []interface{}
	Value interface{}
}

type syncValue struct {
}

func (s *Syncrhoniser) init() {
	s.send, s.newSender = Multicast[*syncMsg]()
}

func (s *Syncrhoniser) SyncReader(r io.Reader) error {
	s.initOnce.Do(s.init)
	dec := gob.NewDecoder(r)

	for {
		msg := new(syncMsg)
		err := dec.Decode(msg)
		if err != nil {
			return err
		}

		value := s.t.Get(msg.Key...)
		if value.syncValue == nil {
			value.syncValue = &syncValue{
				sync: s,
				key:  msg.Key,
			}
			s.t.Set(value, msg.Key...)
		}

		value.mutex.Lock()
		if !value.lastChange.Before(msg.T) {
			value.mutex.Unlock()
			continue
		}

		value.lastChange = msg.T
		value.value = msg.Value

		s.send <- msg
	}
}

func (s *Syncrhoniser) SyncWriter(w io.Writer) error {
	s.initOnce.Do(s.init)
	enc := gob.NewEncoder(w)

	recv, done := s.newSender()
	defer done()

	for msg := range recv {
		err := enc.Encode(msg)
		if err != nil {
			return err
		}
	}

	return nil
}
