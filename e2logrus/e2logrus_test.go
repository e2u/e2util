package e2logrus

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger(&Config{
		Output:              "stdout",
		Level:               "info",
		Format:              "text",
		DisableReportCaller: true,
		DisableColor:        true,
	})
	if log == nil {
		t.Fatal("NewLogger returned nil")
	}
	if log.GetLevel() != logrus.InfoLevel {
		t.Errorf("level = %v, want info", log.GetLevel())
	}

	jsonLog := NewLogger(&Config{
		Output:              "stdout",
		Level:               "debug",
		Format:              "json",
		DisableReportCaller: true,
	})
	if jsonLog.GetLevel() != logrus.DebugLevel {
		t.Errorf("json logger level = %v, want debug", jsonLog.GetLevel())
	}

	fallback := NewLogger(&Config{Level: "not-a-level", Format: "text", DisableReportCaller: true, DisableColor: true})
	if fallback.GetLevel() != logrus.InfoLevel {
		t.Errorf("invalid level should fall back to info, got %v", fallback.GetLevel())
	}
}

func TestCloneLogrus(t *testing.T) {
	orig := NewLogger(&Config{Level: "warn", Format: "text", DisableReportCaller: true, DisableColor: true})
	cloned := CloneLogrus(orig)
	if cloned == orig {
		t.Fatal("CloneLogrus returned the same logger")
	}
	if cloned.GetLevel() != orig.GetLevel() {
		t.Errorf("cloned level = %v, want %v", cloned.GetLevel(), orig.GetLevel())
	}
}

func TestSeqHook(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()
	log.SetOutput(&buf)
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableColors: true})
	log.AddHook(&SeqHook{})
	log.Info("hello")
	if buf.Len() == 0 {
		t.Fatal("expected log output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("seq=")) {
		t.Errorf("expected seq field, got %q", buf.String())
	}
}
