package e2app

import (
	"context"
	"embed"
	"testing"

	"github.com/e2u/e2util/e2db"
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

func Test_cfgFS(t *testing.T) {
	if err := DebugFS(cfgFs, "."); err != nil {
		t.Fatalf("DebugFS: %v", err)
	}
	b, err := cfgFs.ReadFile("dev.toml")
	if err != nil {
		t.Fatalf("read embedded dev.toml: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("embedded dev.toml is empty")
	}
}

func skipIfPostgresUnavailable(t *testing.T) {
	t.Helper()
	cfg := &e2db.Config{
		Driver: "postgres",
		Writer: "host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=database sslmode=disable TimeZone=UTC application_name=e2app-test",
	}
	conn, err := e2db.New(cfg)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	sqlDB, err := conn.RW().DB()
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("postgres not available: %v", err)
	}
}

func TestNew(t *testing.T) {
	skipIfPostgresUnavailable(t)

	ctx := New(context.TODO(), cfgFs)
	if ctx == nil {
		t.Fatal("New returned nil")
	}
	if ctx.Cache == nil {
		t.Fatal("expected cache to be initialized")
	}
	if err := ctx.Cache.Set(context.Background(), "key", "value"); err != nil {
		t.Fatal(err)
	}
	got, err := ctx.Cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(got)

	var ex *Extend
	if err := viper.Unmarshal(&ex); err != nil {
		t.Fatal(err)
	}
	t.Log(ex)
}

func TestAppConfig(t *testing.T) {
	ap := &AppConfig{
		Name: "demo",
		ExtraProps: map[string]any{
			"abc":        "ffefef",
			"ccc":        12345,
			"secret_key": "c2VjcmV0X2tleQo=",
			"tags":       []any{"golang", "viper"},
			"settings":   map[string]any{"debug": true},
			"enabled":    true,
			"ratio":      1.5,
		},
	}
	if ap.Get("abc") != "ffefef" {
		t.Errorf("Get(abc) = %v", ap.Get("abc"))
	}
	if ap.GetString("abc") != "ffefef" {
		t.Errorf("GetString(abc) = %q", ap.GetString("abc"))
	}
	if ap.GetInt("ccc") != 12345 {
		t.Errorf("GetInt(ccc) = %d", ap.GetInt("ccc"))
	}
	if ap.GetFloat("ratio") != 1.5 {
		t.Errorf("GetFloat(ratio) = %v", ap.GetFloat("ratio"))
	}
	if !ap.GetBool("enabled") {
		t.Error("GetBool(enabled) = false")
	}
	if got := ap.GetStringSlice("tags"); len(got) != 2 || got[0] != "golang" {
		t.Errorf("GetStringSlice(tags) = %v", got)
	}
	if got := ap.GetStringMap("settings"); got["debug"] != true {
		t.Errorf("GetStringMap(settings) = %v", got)
	}
	if got := ap.GetBytesFromBase64("secret_key"); string(got) != "secret_key\n" {
		t.Errorf("GetBytesFromBase64(secret_key) = %q", got)
	}
	if ap.Get("missing") != nil {
		t.Errorf("Get(missing) = %v, want nil", ap.Get("missing"))
	}
}
