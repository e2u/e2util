package e2map

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

type Map map[string]any

var lock sync.RWMutex

func (ms Map) Range(fn func(key string, value any)) {
	lock.RLock()
	defer lock.RUnlock()
	for key, value := range ms {
		fn(key, value)
	}
}

func (ms Map) Set(key string, value any) {
	lock.Lock()
	defer lock.Unlock()
	ms[key] = value
}

func (ms Map) Get(key string) (any, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := ms[key]
	return v, ok
}

func (ms Map) KeyExists(key string) bool {
	lock.RLock()
	defer lock.RUnlock()
	_, ok := ms[key]
	return ok
}

func (ms Map) DefaultGet(key string, defVal any) (any, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := ms[key]
	if !ok {
		return defVal, false
	}
	return v, true
}

func (ms Map) DefaultString(key string, defVal string) (string, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := ms.DefaultGet(key, defVal)
	if !ok {
		return defVal, false
	}
	if tv, ok := v.(string); ok {
		return tv, true
	}
	return defVal, false
}

func (ms Map) DefaultInt(key string, defVal int) (int, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := ms.DefaultGet(key, defVal)
	if !ok {
		return defVal, false
	}
	if i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64); err == nil {
		return int(i), true
	} else {
		return defVal, false
	}
}

func (ms Map) DefaultBool(key string, defVal bool) (bool, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := ms.DefaultGet(key, defVal)
	if !ok {
		return defVal, false
	}
	if tv, ok := v.(bool); ok {
		return tv, true
	}
	stv := strings.ToLower(fmt.Sprintf("%v", v))
	if stv == "true" || stv == "1" {
		return true, true
	} else {
		return defVal, false
	}
}

func (ms Map) DecodeBase64Value(key string) ([]byte, error) {
	lock.RLock()
	defer lock.RUnlock()
	if v, ok := ms[key]; ok {
		return base64.StdEncoding.DecodeString(v.(string))
	}
	return nil, errors.New("key not found")
}
