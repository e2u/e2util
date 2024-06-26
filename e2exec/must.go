package e2exec

import (
	"io"

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
	if v, ok := fn.(io.Closer); ok {
		if err := v.Close(); err != nil {
			logrus.Errorf("e2exec.MustClose() error=%v", err)
		}
	}
}

func SilentError(args ...any) {
	if len(args) == 0 {
		return
	}
	if len(args) == 1 && args[0] != nil {
		logrus.Errorf("e2exec.SilentError() error=%v", args[0])
		return
	}
	switch err := args[len(args)-1].(type) {
	case error:
		if err != nil {
			logrus.Errorf("e2exec.SilentError() error=%v", err)
		}
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
