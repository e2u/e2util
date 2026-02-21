package e2sync

import (
	"sort"
	"sync"
	"sync/atomic"
)

// SyncMapLen 计算 sync.Map 的大小
func SyncMapLen(lock *sync.RWMutex, m *sync.Map) uint64 {
	lock.RLock()
	defer lock.RUnlock()
	var ri uint64
	m.Range(func(key, val any) bool {
		atomic.AddUint64(&ri, 1)
		return true
	})
	return ri
}

// SyncMapSortStringKeys 返回指定 sync.Map 的字符串key
func SyncMapSortStringKeys(lock *sync.RWMutex, m *sync.Map) []string {
	lock.RLock()
	defer lock.RUnlock()
	var keys []string
	m.Range(func(key, val any) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	return keys
}

// Add
// increase the value of the key in the map and return new value and true if the exists value of the key is same type.
// if the value exist in m by the key but the type not match with inc than restore the int to map m
func Add[T int | int64 | uint | uint64](m *sync.Map, key any, inc T) (T, bool) {
	for {
		actual, loaded := m.LoadOrStore(key, inc)
		if !loaded {
			return inc, false
		}
		if v, ok := actual.(T); ok {
			newValue := v + inc
			if m.CompareAndSwap(key, actual, any(newValue)) {
				return newValue, true
			}
			continue
		}
		m.Store(key, inc)
		return inc, false
	}
}
