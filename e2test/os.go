package e2test

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/e2u/e2util/e2os"
	"github.com/sirupsen/logrus"
)

func Chroot() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("get current dir error=%v", err)
		return
	}
	for flag.Lookup("test.v") != nil && pwd != "/" {
		if e2os.FileExists(filepath.Join(pwd, "go.mod")) {
			if err := os.Chdir(pwd); err != nil {
				logrus.Errorf("chdir error=%v", err)
			}
			break
		}
		pwd = filepath.Dir(pwd)
	}
}
