package strata_test

import (
	"testing"

	"github.com/stewi1014/strata"
)

func TestGraphAssignments(t *testing.T) {
	var g strata.Graph

	if n := g.Len(); n != 0 {
		t.Errorf("Len() of empty Graph is %v, wanted 0", n)
	}

	g = strata.MakeGraph(2)

	g.Set(0, 1)
	if !g.Get(0, 1) {
		t.Errorf("Get(0, 1) returned false after call to Set(0, 1), wanted true")
	}

	if g.Get(1, 0) {
		t.Errorf("Get(1, 0) returned true after call to Set(0, 1); directionality is not respected")
	}

	g.Unset(0, 1)
	if g.Get(0, 1) {
		t.Errorf("Get(0, 1) returned true after call to Clear(0, 1), wanted false")
	}
}
