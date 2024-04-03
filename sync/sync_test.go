package sync_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stewi1014/strata/sync"
)

func changeWait[T any](v *sync.Value[T]) chan T {
	c := make(chan T)
	v.OnChange(func(value any, t time.Time) {
		c <- value.(T)
	})
	return c
}

type TestStruct struct {
	Value1 sync.Value[string]
	Value2 sync.Value[int]
}

func TestSync(t *testing.T) {
	ctx, done := context.WithCancelCause(context.Background())
	conn1, conn2 := net.Pipe()

	go func() {
		s := new(sync.Sync)
		go func() {
			err := s.Sync(conn1)
			if err != nil {
				done(err)
			}
		}()

		value2 := new(sync.Value[string])
		value2.Set("hello")

		s.Register(value2, "value2")

		valueStruct := &TestStruct{}
		valueStruct.Value1.Set("hello struct")
		s.RegisterStruct(valueStruct)

		valueStruct.Value2.Set(42)
	}()

	go func() {
		defer done(context.Canceled)

		s := new(sync.Sync)
		go func() {
			err := s.Sync(conn2)
			if err != nil {
				done(err)
			}
		}()

		value2 := new(sync.Value[string])
		value2Change := changeWait(value2)
		s.Register(value2, "value2")

		valueStruct := &TestStruct{}
		valueStructValue1 := changeWait(&valueStruct.Value1)
		valueStructValue2 := changeWait(&valueStruct.Value2)
		s.RegisterStruct(valueStruct)

		v2 := <-value2Change
		if v2 != "hello" {
			t.Errorf(v2)
		}

		sv1 := <-valueStructValue1
		if sv1 != "hello struct" {
			t.Errorf(sv1)
		}

		sv2 := <-valueStructValue2
		if sv2 != 42 {
			t.Errorf("%v", sv2)
		}
	}()

	<-ctx.Done()
	err := context.Cause(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
