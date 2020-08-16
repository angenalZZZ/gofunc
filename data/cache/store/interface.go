package store

import "time"

// Interface is the interface for all available stores
type Interface interface {
	Get(key string) (interface{}, error)
	TTL(key string) (time.Duration, error)
	Set(key string, value interface{}, options *Options) error
	Delete(key string) error
	Invalidate(options InvalidateOptions) error
	Clear() error
	GetType() string
}
