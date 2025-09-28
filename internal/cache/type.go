package cache

import "context"

type Client interface {
	SetValue(ctx context.Context, key string, value any) error
	GetValue(ctx context.Context, key string) (string, error)
}
