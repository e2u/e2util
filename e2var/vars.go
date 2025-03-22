package e2var

import "reflect"

func MustStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func P[T any](i T) *T {
	return &i
}

func NeverNullPoint[T any](i T, defVal T) *T {
	v := reflect.ValueOf(i)
	if !v.IsValid() || v.IsZero() {
		return P(defVal)
	}
	return P(i)
}

func NeverNull[T any](i T, defVal T) T {
	return *NeverNullPoint(i, defVal)
}

func IfElse[T comparable, R any](v1, v2 T, r1, r2 R) R {
	if v1 == v2 {
		return r1
	}
	return r2
}

func IfElseFunc[T comparable](v1, v2 T, f1 func(), f2 func()) {
	if v1 == v2 {
		f1()
	} else {
		f2()
	}
}

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

func ValueOrDefault[T any](input T, defVal T) T {
	return *NeverNullPoint(input, defVal)
}

func ExpectOrDefault[T any, T1 any](input T, defVal T1) (T1, bool) {
	if v, ok := any(input).(T1); ok {
		return v, true
	}
	return defVal, false
}
