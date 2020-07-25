package store

// Interface is the interface for all available stores
type Interface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, options *Options) error
	TTL(key string) int64
	Delete(key string) error
	Invalidate(options InvalidateOptions) error
	Clear() error
	GetType() string
}
