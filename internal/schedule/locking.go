package schedule

import "sync"

var RunningWorkflows = NewSafeMapMutex()

type SafeMapRWMutex struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewSafeMapMutex() *SafeMapRWMutex {
	return &SafeMapRWMutex{
		data: make(map[string]string),
	}
}

func (sm *SafeMapRWMutex) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

func (sm *SafeMapRWMutex) Set(key string, value string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

func (sm *SafeMapRWMutex) Get(key string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, ok := sm.data[key]

	return val, ok
}
