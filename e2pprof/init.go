package e2pprof

import (
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
)

var (
	PprofPort int
)

func init() {
	go func() {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			logrus.Errorf("make tcp listen error:%v", err)
			return
		}
		PprofPort = listener.Addr().(*net.TCPAddr).Port
		logrus.Infof("pprof port: %v", PprofPort)
		if err := http.Serve(listener, nil); err != nil {
			logrus.Infof("run pprof error:%v", err)
			return
		}
	}()
}
