package strata

import (
	"sync"
)

type Value[T any] struct {
	mutex sync.Mutex

	v        T
	updatees []*func(T)
}

func (value *Value[T]) OnChange(onChange func(T)) (done func()) {
	value.mutex.Lock()
	defer value.mutex.Unlock()

	ptr := &onChange
	value.updatees = append(value.updatees, ptr)

	return func() {
		for i, f := range value.updatees {
			if f == ptr {
				copy(value.updatees[i:], value.updatees[i+1:])
				value.updatees = value.updatees[:len(value.updatees)-1]
				return
			}
		}
	}
}

func (value *Value[T]) Get() T {
	value.mutex.Lock()
	defer value.mutex.Unlock()
	return value.v
}

func (value *Value[T]) Set(v T) {
	value.mutex.Lock()
	defer value.mutex.Unlock()

	value.v = v
	for _, f := range value.updatees {
		go (*f)(v)
	}
}
