package cache

import (
	"context"
)

type Client interface {
	Set(ctx context.Context, key, mime string, data interface{}, isFile bool) error
	Get(ctx context.Context, key string) ([]byte, string, bool)
	Delete(ctx context.Context, key string) error
}
