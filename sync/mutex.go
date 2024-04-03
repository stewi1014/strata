package sync

import "sync"

type MutexTrack struct {
	Lock   bool
	Rlocks int
}

func (m *MutexTrack) Assert() {
	if m.Lock {
		panic("lock still held")
	}

	if m.Rlocks > 0 {
		panic("read locks still held")
	}
}

type DebugRMMutex struct {
	rwmutex sync.RWMutex
}

func (d *DebugRMMutex) Lock(t *MutexTrack) {
	d.rwmutex.Lock()
	t.Lock = true
}

func (d *DebugRMMutex) Unlock(t *MutexTrack) {
	d.rwmutex.Unlock()
	t.Lock = false
}

func (d *DebugRMMutex) RLock(t *MutexTrack) {
	d.rwmutex.RLock()
	t.Rlocks++
}

func (d *DebugRMMutex) RUnlock(t *MutexTrack) {
	d.rwmutex.RUnlock()
	t.Rlocks--
}
