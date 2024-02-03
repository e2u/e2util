package e2map

import (
	"sync"
)

func Keys[K comparable, V any](l *sync.RWMutex, m map[K]V) []K {
	l.RLock()
	defer l.RUnlock()
	var rs []K
	for k := range m {
		rs = append(rs, k)
	}
	return rs
}

func Values[K comparable, V any](l *sync.RWMutex, m map[K]V) []V {
	l.RLock()
	defer l.RUnlock()
	var rs []V
	for _, v := range m {
		rs = append(rs, v)
	}
	return rs
}

func DefaultValue[K comparable, V any](l *sync.RWMutex, m map[K]V, key K, defaultValue V) V {
	l.Lock()
	defer l.Unlock()
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

func LoadOrStore[K comparable, V any](l *sync.RWMutex, m map[K]V, key K, value V) (V, bool) {
	l.Lock()
	defer l.Unlock()
	if v, ok := m[key]; ok {
		return v, ok
	}
	m[key] = value
	return value, false
}
