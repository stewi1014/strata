package strata

import (
	"sync"
	"sync/atomic"
)

// Tree holds values in a tree-like structure.
// Every element in the tree stores a value and any number of branches.
//
// A Tree must not be copied after first use.
//
// It is thread safe, but also blocking.
// It utilises sync.RWMutex and methods only hold the lock for one element at a time.
type Tree[T any] struct {
	v atomic.Pointer[T]

	branchesMutex sync.RWMutex
	branches      map[interface{}]*Tree[T]
}

// touch returns the branch at the given key, creating it if neccecary,
// while attempting to only hold a read lock.
//
// *Mutex must not be held.*
//
// Motivation is the duplicate code polluting other functions
// that's required to deal with the transition from RLock to Lock.
func (t *Tree[T]) touch(key interface{}) *Tree[T] {
	// try get to the branch with only a read lock
	t.branchesMutex.RLock()
	branch := t.branches[key]
	t.branchesMutex.RUnlock()
	if branch != nil {
		return branch
	}

	// we need to create the branch.
	// The map might be nil, the next branch doesn't exist,
	// or maybe the branch has spontaneously appeared and we don't need to create it.
	//
	// We lost the lock during the RLock -> Lock transition,
	// and anything could have happened.
	t.branchesMutex.Lock() // start from scratch with full lock.
	branch = t.branches[key]
	if branch != nil {
		// another thread did create the branch
		t.branchesMutex.Unlock()
		return branch
	}

	if t.branches == nil {
		t.branches = make(map[interface{}]*Tree[T])
	}

	branch = new(Tree[T])
	t.branches[key] = branch
	t.branchesMutex.Unlock()
	return branch
}

// Set assigns the value of the given location.
//
// It always succeeds, internally creating any branches required to reach the given key.
func (t *Tree[T]) Set(value T, key ...interface{}) {
	if len(key) == 0 {
		t.v.Store(&value)
		return
	}

	t.touch(key[0]).Set(value, key[1:]...)
}

// Get returns the value at the given key.
//
// If the key does not exist it returns nil.
func (t *Tree[T]) Get(key ...interface{}) *T {
	return t.Branch(key...).v.Load()
}

// Branch returns the tree at the given key.
//
// If the key does not exist it returns nil.
func (t *Tree[T]) Branch(key ...interface{}) *Tree[T] {
	if t == nil {
		return nil
	}

	if len(key) == 0 {
		return t
	}

	t.branchesMutex.RLock()
	next := t.branches[key[0]]
	t.branchesMutex.RUnlock()
	return next.Branch(key[1:]...)
}

// Prune removes the whole tree at key,
// returning the removed subtree.
//
// If the key does not exist it does nothing and returns nil.
//
// If Prune is called with no key, it does nothing and returns itself.
func (t *Tree[T]) Prune(key ...interface{}) *Tree[T] {
	if len(key) == 0 {
		return t
	}

	if t == nil {
		return nil
	}

	if len(key) == 1 {
		t.branchesMutex.Lock()
		subtree := t.branches[key[0]]
		delete(t.branches, key[0])
		t.branchesMutex.Unlock()
		return subtree
	}

	t.branchesMutex.RLock()
	next := t.branches[key[0]]
	t.branchesMutex.RUnlock()
	return next.Prune(key[1:]...)
}

// Graft merges two trees together.
//
// If two elements share the same key the graft takes priority,
// overwriting values, with branches being merged.
func (t *Tree[T]) Graft(graft *Tree[T], key ...interface{}) {
	if t == graft {
		return
	}

	if len(key) == 0 {
		t.branchesMutex.Lock()
		graft.branchesMutex.RLock()

		t.v.Store(graft.v.Load())

		if len(t.branches) == 0 {
			t.branches = graft.branches
			graft.branchesMutex.RUnlock()
			t.branchesMutex.Unlock()
			return
		}

		for k, v := range graft.branches {
			existing := t.branches[k]
			if existing == nil {
				t.branches[k] = v
			} else {
				existing.Graft(v)
			}
		}

		graft.branchesMutex.RUnlock()
		t.branchesMutex.Unlock()

		return
	}

	t.touch(key[0]).Graft(graft, key[1:]...)
}

// Range iterates over branches in the tree.
// It does not recurse.
//
// The function can modify branches given to it,
// but cannot modify the parent tree as the read mutex needs to be held.
func (t *Tree[T]) Range(f func(key interface{}, branch *Tree[T])) {
	if t == nil {
		return
	}

	t.branchesMutex.RLock()
	defer t.branchesMutex.RUnlock()

	for key, branch := range t.branches {
		f(key, branch)
	}
}
