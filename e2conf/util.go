package e2conf

import (
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/**
[StorageDarwin]
PhotosDir = "/Users/tk/Pictures/Wallpapers"
BadgerDir = "./badger"

[StorageLinux]
PhotosDir = "/mnt/s1/images"
BadgerDir = "./badger"

badgerDir, ok := cfg.GetStringMapStringByOS("Storage")["badgerdir"]

*/

func getStringMapStringByOS(v *viper.Viper, key string) map[string]string {
	mapKey := key + "_Darwin"
	switch runtime.GOOS {
	case "linux":
		mapKey = key + "_Linux"
	case "darwin":
		mapKey = key + "_Darwin"
	case "windows":
		logrus.Errorf("windows not yet supported")
	default:
		logrus.Errorf("Unknown operating system.")
	}
	return v.GetStringMapString(mapKey)
}
