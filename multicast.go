package strata

import "sync"

func Multicast[T any]() (send chan<- T, newListener func() (recv <-chan T, done func())) {
	s := make(chan T)

	var mutex sync.Mutex
	var listeners []chan T

	go func() {
		for msg := range s {
			mutex.Lock()
			for _, listener := range listeners {
				listener <- msg
			}
			mutex.Unlock()
		}

		mutex.Lock()
		for _, listener := range listeners {
			close(listener)
		}
		listeners = nil
		mutex.Unlock()
	}()

	newListener = func() (recv <-chan T, done func()) {
		r := make(chan T, 20)

		mutex.Lock()
		listeners = append(listeners, r)
		mutex.Unlock()

		done = func() {
			mutex.Lock()
			for i, listener := range listeners {
				if listener == r {
					copy(listeners[i:], listeners[i+1:])
					listeners = listeners[:len(listeners)-1]
					break
				}
			}
			mutex.Unlock()

			close(r)
		}

		return r, done
	}

	return s, newListener
}
