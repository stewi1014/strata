package sync

import (
	"reflect"
	"sync"
	"time"
)

// Valuelike provides a method for contexts where generics are impossible
// (struct methods, reflection) to interact with Value.
type Valuelike interface {
	// GetAt returns the current value and the time at which it was set.
	GetAt() (any, time.Time)

	// SetAt assigns the value,
	// recording the time at which the value was set as the given time.
	//
	// If the given time is before when the value was last set,
	// it does nothing.
	SetAt(any, time.Time)

	// OnChange calls the given function when the value changes.
	//
	// The function is called with the value and the new time.
	OnChange(func(v any, t time.Time)) (remove func())
}

// ValueOf returns a Value of the given type
// with its value set.
//
// It is equivilent to calling Set on the zero value.
func ValueOf[T any](value T) *Value[T] {
	v := &Value[T]{}
	v.Set(value)
	return v
}

// Value contains a value, wrapping it with
// a mutex, update hooks and time.
//
// Some functions return or accept any instead of T.
// This is to allow Value to conform to a non-generic interface (Valuelike).
type Value[T any] struct {
	mutex sync.Mutex
	hooks []*func(any, time.Time)
	t     time.Time
	v     T
}

var valuelikeType = reflect.TypeOf(new(Valuelike)).Elem()

var _ Valuelike = &Value[any]{}

// GetAt returns the current value and the time the value was set.
func (v *Value[T]) GetAt() (any, time.Time) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.v, v.t
}

// Get returns the value.
func (v *Value[T]) Get() T {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.v
}

// Set sets the value, recording the curent time.
func (v *Value[T]) Set(value T) {
	v.setAt(value, time.Now())
}

// SetAt sets the value at some point in the past.
// If the value has been updated since t, the value
// is not changed and hooks are not called.
func (v *Value[T]) SetAt(value any, t time.Time) {
	v.setAt(value.(T), t)
}

func (v *Value[T]) setAt(value T, t time.Time) {
	v.mutex.Lock()

	if !v.t.Before(t) {
		// value has been updated since,
		// keep the newer value
		v.mutex.Unlock()
		return
	}

	v.t = t
	v.v = value

	// copied to prevent the possibility of mutex deadlock.
	// (unreasonable to expect callback to never touch this value)
	hooks := make([]*func(v any, t time.Time), len(v.hooks))
	copy(hooks, v.hooks)
	v.mutex.Unlock()

	for _, f := range hooks {
		(*f)(value, t)
	}
}

// OnChange adds a hook that is called whevener the value is changed.
//
// It does not guarantee that currently executing calls to set
// won't still call the callback even after remove returns.
func (v *Value[T]) OnChange(f func(v any, t time.Time)) (remove func()) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	ptr := &f
	v.hooks = append(v.hooks, ptr)
	return func() {
		v.mutex.Lock()
		defer v.mutex.Unlock()

		del(&v.hooks, ptr)
	}
}
