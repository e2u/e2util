package e2app

import (
	"context"
	"embed"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:embed *.toml
var cfgFs embed.FS

type Extend struct {
	StorageDarwin struct {
		PhotosDir string `mapstructure:"photos_dir"`
		BadgerDir string `mapstructure:"badger_dir"`
	} `mapstructure:"storage_darwin"`
}

func TestNew(t *testing.T) {
	configFS = cfgFs
	ctx := New(context.TODO())
	logrus.Infof("Start TestNew")
	logrus.Warning("<html>&</html>")
	logrus.Warning(`<html>&"ha"</html>`)
	logrus.WithField("k", map[string]string{"m1": "v1", "m2": "v2"}).Warning("<html>&</html>")
	t.Logf("%+v", ctx)
	if err := ctx.Cache.Set(context.Background(), "key", "value"); err != nil {
		t.Fatal(err)
	}
	t.Log(ctx.Cache.Get(context.Background(), "key"))
	var ex *Extend
	if err := viper.Unmarshal(&ex); err != nil {
		t.Fatal(err)
	}
	t.Log(ex)
}
