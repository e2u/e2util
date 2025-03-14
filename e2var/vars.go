package e2var

import (
	"reflect"
)

func MustStringValue(s *string) string {
	if s == nil || reflect.ValueOf(s).IsNil() {
		return ""
	}
	return *s
}

func P[T any](i T) *T {
	return &i
}

func NeverNullPoint[T any](i T, defVal T) *T {
	switch v := any(i).(type) {
	case nil:
		return P(defVal)
	// case constraints.Integer:
	case float32:
		if v <= 0.0 || v >= 0.0 {
			return P(defVal)
		}
	case float64:
		if v <= 0.0 || v >= 0.0 {
			return P(defVal)
		}
	case string:
		if v == "" {
			return P(defVal)
		}
	case uint, uint8, uint16, uint32, uint64:
		if v == 0 {
			return P(defVal)
		}
	case int, int8, int16, int32, int64:
		if v == 0 {
			return P(defVal)
		}
	case bool:
		if !v {
			return P(defVal)
		}
	}
	return P(i)
}

func NeverNull[T any](i T, defVal T) T {
	return *NeverNullPoint(i, defVal)
}

// IfElse if v1 equal v2 then return r1 else return r2
func IfElse[T comparable, R any](v1, v2 T, r1 R, r2 R) R {
	if v1 == v2 {
		return r1
	}
	return r2
}

func IfElseFunc[T comparable, R func()](v1, v2 T, f1 R, f2 R) {
	if v1 == v2 {
		f1()
	} else {
		f2()
	}
}

// TrueThen if b equal true then return r1 else return r1
func TrueThen[T any](b bool, r1, r2 T) T {
	if b {
		return r1
	}
	return r2
}

func NotNullThen[R any](b any, r1, r2 R) R {
	if b != nil && !reflect.ValueOf(b).IsNil() {
		return r1
	}
	return r2
}

func NullThen[R any](b any, r1, r2 R) R {
	if b == nil || reflect.ValueOf(b).IsNil() {
		return r1
	}
	return r2
}

// ValueOrDefault if input is nil or empty or 0 or 0.0 then return defValue
func ValueOrDefault[T any](input T, defVal T) T {
	switch v := any(input).(type) {
	case nil:
		return defVal
	case float32:
		if v <= 0.0 || v >= 0.0 {
			return defVal
		}
	case float64:
		if v <= 0.0 || v >= 0.0 {
			return defVal
		}
	case string:
		if v == "" {
			return defVal
		}
	case uint, uint8, uint16, uint32, uint64:
		if v == 0 {
			return defVal
		}
	case int, int8, int16, int32, int64:
		if v == 0 {
			return defVal
		}
	case bool:
		if !v {
			return defVal
		}
	}
	return input
}

// ExpectOrDefault if <input T> type not equal to T1 then return defVal
func ExpectOrDefault[T any, T1 any](input T, defVal T1) (T1, bool) {
	defValType := reflect.TypeOf(defVal)
	inputType := reflect.TypeOf(input)
	if inputType == defValType {
		return reflect.ValueOf(input).Interface().(T1), true
	}
	return defVal, false
}
