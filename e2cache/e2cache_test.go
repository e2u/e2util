package e2cache

import (
	"context"
	"testing"
)

func TestMemoryCache(t *testing.T) {
	c := New(&Config{Type: "memory"})
	if c == nil || c.Cache == nil {
		t.Fatal("expected memory cache")
	}
	ctx := context.Background()
	if err := c.Set(ctx, "k", "v"); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if got != "v" {
		t.Errorf("Get = %v, want v", got)
	}
	if err := c.Delete(ctx, "k"); err != nil {
		t.Fatal(err)
	}
}

func TestFakeCache(t *testing.T) {
	c := New(&Config{Type: "fake"})
	ctx := context.Background()
	if err := c.Set(ctx, "k", "v"); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("fake Get = %v, want nil", got)
	}
	if typ := NewFakeCacheStore[any]().GetType(); typ != "fake" {
		t.Errorf("GetType = %q", typ)
	}
}
