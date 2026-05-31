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

func (sm *SafeMapRWMutex) TryAcquire(key, value string) (string, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, ok := sm.data[key]
	if ok {
		return existing, false
	}

	sm.data[key] = value
	return "", true
}
