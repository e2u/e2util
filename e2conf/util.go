package e2conf

import (
	"maps"
	"runtime"
	"strings"

	"github.com/e2u/e2util/e2map"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/**
[storage_darwin]
photos_dir = "/Volumes/r1/images"
badger_dir = "./badger"

[storage_linux]
photos_dir = "/mnt/s1/images"
dadger_dir = "./badger"


*/

func getStringMapByOS(v *viper.Viper, key string) e2map.Map {
	mapKey := key + "_darwin"
	switch strings.ToLower(runtime.GOOS) {
	case "linux":
		mapKey = key + "_linux"
	case "darwin":
		mapKey = key + "_darwin"
	case "windows":
		logrus.Errorf("windows not yet supported")
	default:
		logrus.Errorf("Unknown operating system.")
	}
	r := make(e2map.Map)
	maps.Copy(r, v.GetStringMap(mapKey))
	return r
}
