package gdnative

import (
	"sync"
)

type SyncMap[K comparable, V any] struct {
	sync.RWMutex
	internal map[K]V
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		internal: make(map[K]V),
	}
}

func (m *SyncMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	result, ok := m.internal[key]
	m.RUnlock()
	return result, ok
}

func (m *SyncMap[K, V]) Set(key K, value V) {
	m.Lock()
	m.internal[key] = value
	m.Unlock()
}

func (m *SyncMap[K, V]) Delete(key K) {
	m.Lock()
	delete(m.internal, key)
	m.Unlock()
}

func (m *SyncMap[K, V]) Keys() []K {
	m.Lock()
	ks := make([]K, 0, len(m.internal))
	for k := range m.internal {
		ks = append(ks, k)
	}
	m.Unlock()
	return ks
}

func (m *SyncMap[K, V]) HasKey(key K) bool {
	m.Lock()
	_, ok := m.internal[key]
	m.Unlock()
	return ok
}

func (m *SyncMap[K, V]) Values() []V {
	m.Lock()
	vs := make([]V, 0, len(m.internal))
	for _, v := range m.internal {
		vs = append(vs, v)
	}
	m.Unlock()
	return vs
}
