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
// It is only possible to traverse down the tree, not up.
//
// Don't create loops. Loops are not trees. Tree will recurse infintiely.
//
// It is thread safe, but also blocking.
// It utilises sync.RWMutex and methods only hold the lock for one element at a time.
type Tree struct {
	v atomic.Pointer[interface{}]

	onChange []*func(*Tree)

	branchesMutex sync.RWMutex
	branches      map[interface{}]*Tree
}

// touch returns the branch at the given key, creating it if neccecary,
// while attempting to only hold a read lock.
//
// *Mutex must not be held.*
//
// Motivation is the duplicate code polluting other functions
// that's required to deal with the transition from RLock to Lock.
func (t *Tree) touch(key interface{}) *Tree {
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
		t.branches = make(map[interface{}]*Tree)
	}

	branch = new(Tree)
	t.branches[key] = branch
	t.branchesMutex.Unlock()
	return branch
}

// Get returns the value at the given key.
//
// If the key does not exist it returns nil.
func (t *Tree) Get(key ...interface{}) interface{} {
	ptr := t.Branch(key...).v.Load()
	if ptr == nil {
		ptr = new(interface{})
	}
	return *ptr
}

// Set assigns the value of the given location.
//
// It always succeeds, internally creating any branches required to reach the given key.
func (t *Tree) Set(value interface{}, key ...interface{}) {
	if len(key) == 0 {
		t.v.Store(&value)
		return
	}

	t.touch(key[0]).Set(value, key[1:]...)
}

// OnChange registers a function that is called when the value is modified.
func (t *Tree) OnChange(onValue func(*Tree), key ...interface{}) {

}

// On

// Branch returns the tree at the given key.
//
// If the key does not exist it returns nil.
func (t *Tree) Branch(key ...interface{}) *Tree {
	if len(key) == 0 {
		return t
	}

	if t == nil {
		return nil
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
func (t *Tree) Prune(key ...interface{}) *Tree {
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
func (t *Tree) Graft(graft *Tree, key ...interface{}) {
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
func (t *Tree) Range(f func(key interface{}, branch *Tree)) {
	if t == nil {
		return
	}

	t.branchesMutex.RLock()
	defer t.branchesMutex.RUnlock()

	for key, branch := range t.branches {
		f(key, branch)
	}
}
