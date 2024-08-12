package e2conf

import (
	"reflect"
	"testing"

	"github.com/e2u/e2util/e2conf/etc"
)

func Test_GetStringMapByOS(t *testing.T) {
	cfg := New(&InitConfigInput{
		Env:        "dev",
		ConfigFs:   etc.Fs,
		ConfigName: "example-app-dev",
	})

	t.Run("DefaultGet", func(t *testing.T) {
		photosDir, _ := cfg.GetStringMapByOS("storage").DefaultGet("photos_dir", "/tmp")
		t.Log(photosDir)
	})

	t.Run("DefaultString", func(t *testing.T) {
		rs, ok := cfg.GetStringMapByOS("storage").DefaultString("photos_dir", "/tmp")
		if !ok {
			t.Fatalf("return type is no string")
		}
		if !reflect.TypeOf(rs).ConvertibleTo(reflect.TypeOf("")) {
			t.Fatalf("return type is not string")
		}
		t.Log(rs)
	})
}
