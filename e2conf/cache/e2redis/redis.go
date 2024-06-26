package e2redis

import (
	"context"
	"fmt"
	"net/url"

	"github.com/e2u/e2util/e2strconv"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Writer string `mapstructure:"writer"` // "redis://127.0.0.1:14023?db=0"
	Reader string `mapstructure:"reader"` // "redis://127.0.0.1:14023?db=0"
}

type Client struct {
	RW *redis.Client
	RO *redis.Client
}

func New(cfg *Config) *Client {
	return NewWithContext(context.TODO(), cfg)
}
func NewWithContext(ctx context.Context, cfg *Config) *Client {
	cli := &Client{}
	if cfg.Writer != "" {
		if c, err := connect(ctx, cfg.Writer); err == nil {
			cli.RW = c
		}
	}

	if cfg.Reader != "" {
		if c, err := connect(ctx, cfg.Reader); err == nil {
			cli.RO = c
		}
	}

	return cli
}

func connect(ctx context.Context, endpoint string) (*redis.Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint empty")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	opt := &redis.Options{
		Addr: u.Host,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			if v, err := cn.Ping(ctx).Result(); err != nil || v != "PONG" {
				logrus.Errorf("redis %s ping error=%v", u.Host, err)
				return err
			}
			return nil
		},
	}

	if u.User != nil {
		opt.Username = u.User.Username()
		opt.Password, _ = u.User.Password()
	}

	if v := u.Query().Get("db"); v != "" {
		opt.DB = e2strconv.MustParseInt(v)
	}

	cli := redis.NewClient(opt)

	return cli, nil
}
