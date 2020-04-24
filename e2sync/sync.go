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
	m.Range(func(key, val interface{}) bool {
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
	m.Range(func(key, val interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	return keys
}
