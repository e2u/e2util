package e2exec

import (
	"reflect"
)

func TrueThen(b bool, trueFunc, falseFunc func()) func() {
	if b {
		return trueFunc
	}
	return falseFunc
}

func TrueThenExec(b bool, trueFunc, falseFunc func() any) any {
	if b {
		return trueFunc()
	}
	return falseFunc()
}

func TrueThenFunc[T func()](b bool, f1, f2 T) {
	if b {
		f1()
	} else {
		f2()
	}
}

func NotNullThenFunc[R func()](b any, f1, f2 R) {
	if b != nil && !reflect.ValueOf(b).IsNil() {
		f1()
	} else {
		f2()
	}
}

func NullThenFunc[R func()](b any, f1, f2 R) {
	if b == nil || reflect.ValueOf(b).IsNil() {
		f1()
	} else {
		f2()
	}
}
