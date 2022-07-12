package e2io

import (
	"io"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func MustReadAll(r io.Reader) []byte {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		logrus.Errorf("read all error=%v", err)
		return nil
	}
	return b
}

func MustReadAllAsString(r io.Reader) string {
	return string(MustReadAll(r))
}

func MustReadAllAndClose(r io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		logrus.Errorf("read all error=%v", err)
		return nil
	}
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			logrus.Errorf("close read all error=%v", err)
		}
	}(r)
	return b
}

func MustReadAllAsStringAndClose(r io.ReadCloser) string {
	return string(MustReadAllAndClose(r))
}

func MustReadFile(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Errorf("read file error=%v", err)
		return nil
	}
	return b
}
