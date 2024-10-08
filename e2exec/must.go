package e2exec

import (
	"io"
	"runtime"

	"github.com/sirupsen/logrus"
)

func Must[T any](v T, err error) T {
	if err != nil {
		logrus.Errorf("e2exec.Must() error=%v", err)
	}
	return v
}

func Must2[T any, T2 any](v T, v2 T2, err error) (T, T2) {
	if err != nil {
		logrus.Errorf("e2exec.Must() error=%v", err)
	}
	return v, v2
}

func MustClose(fn io.Closer) {
	if err := fn.Close(); err != nil {
		logrus.Errorf("e2exec.MustClose() error=%v", err)
	}
}

func SilentError(args ...any) {
	if len(args) == 0 {
		return
	}

	if len(args) == 1 && args[0] != nil {
		if err, ok := args[0].(error); ok && err != nil {
			logrus.Errorf("e2exec.SilentError() error=%v", err)
		}
		return
	}

	pc, filename, line, _ := runtime.Caller(1)

	switch err := args[len(args)-1].(type) {
	case error:
		if err != nil {
			logrus.Errorf("e2exec.SilentError() error=%v in %s[%s:%d]", err, runtime.FuncForPC(pc).Name(), filename, line)
		}
	case nil:
		return
	default:
		logrus.Warn("e2exec.SilentError() last parameter not an error type")
	}
}

func SilentError_old(args ...any) {
	if len(args) == 0 {
		return
	}

	if len(args) == 1 && args[0] != nil && args[0].(error) != nil {
		logrus.Errorf("e2exec.SilentError() error=%v", args[0])
		return
	}

	pc, filename, line, _ := runtime.Caller(1)

	switch err := args[len(args)-1].(type) {
	case error:
		if err != nil {
			logrus.Errorf("e2exec.SilentError() error=%v in %s[%s:%d]", err, runtime.FuncForPC(pc).Name(), filename, line)
		}
	case nil:
		return
	default:
		logrus.Warn("e2exec.SilentError() last parameter not a error type")
	}
}

func OnlyError(args ...any) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) == 1 && args[0] != nil {
		logrus.Errorf("e2exec.OnlyError() error=%v", args[0])
		if err, ok := args[0].(error); ok {
			return err
		}
	}
	if err, ok := args[len(args)-1].(error); ok {
		return err
	}
	return nil
}
