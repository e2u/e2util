package e2cache

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	redisstore "github.com/eko/gocache/store/redis/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// type StoreInterface interface {
//	Get(ctx context.Context, key any) (any, error)
//	GetWithTTL(ctx context.Context, key any) (any, time.Duration, error)
//	Set(ctx context.Context, key any, value any, options ...Option) error
//	Delete(ctx context.Context, key any) error
//	Invalidate(ctx context.Context, options ...InvalidateOption) error
//	Clear(ctx context.Context) error
//	GetType() string
//}

type FakeCacheStore[T any] struct {
}

func NewFakeCacheStore[T any]() *FakeCacheStore[T] {
	return &FakeCacheStore[T]{}
}
func (fc *FakeCacheStore[T]) Get(ctx context.Context, key any) (any, error) {
	var zero T
	return zero, nil
}
func (fc *FakeCacheStore[T]) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	var zero T
	return zero, 0, nil
}
func (fc *FakeCacheStore[T]) Set(ctx context.Context, key any, value any, options ...store.Option) error {
	return nil
}

func (fc *FakeCacheStore[T]) Delete(ctx context.Context, key any) error {
	return nil
}

func (fc *FakeCacheStore[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return nil
}
func (fc *FakeCacheStore[T]) Clear(ctx context.Context) error {
	return nil
}
func (fc *FakeCacheStore[T]) GetType() string {
	return "fake"
}

// GetType() string

type Config struct {
	Enable bool   `mapstructure:"enable"`
	Type   string `mapstructure:"type"`
	Dsn    string `mapstructure:"dsn"`
}

type Connect struct {
	Enable bool
	Err    error
	*cache.Cache[any]
}

func New(cfg *Config) *Connect {
	c := &Connect{}
	switch cfg.Type {
	case "redis":
		dsn := cfg.Dsn
		logrus.Infof("using redis dsn: %s", dsn)
		opts, err := redis.ParseURL(dsn)
		if err != nil {
			c.Err = err
			return c
		}
		cli := redis.NewClient(opts)
		redisStore := redisstore.NewRedis(cli)
		c.Cache = cache.New[any](redisStore)
	case "memory":
		logrus.Infof("using memory cache")
		goCache := gocachestore.NewGoCache(gocache.New(gocache.NoExpiration, gocache.NoExpiration))
		c.Cache = cache.New[any](goCache)
	default:
		logrus.Infof("using fake cache")
		c.Cache = cache.New[any](NewFakeCacheStore[any]())
	}

	return c
}
