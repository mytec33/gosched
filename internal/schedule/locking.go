package schedule

import "sync"

var RunningWorkflows = NewSafeMap()

type SafeMap struct {
	mu   sync.Mutex
	data map[string]string
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]string),
	}
}

func (sm *SafeMap) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

func (sm *SafeMap) TryAcquire(key, value string) (string, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, ok := sm.data[key]
	if ok {
		return existing, false
	}

	sm.data[key] = value
	return "", true
}
