package main

import "context"

type cacheServiceImpl struct{}

var defaultService *cacheServiceImpl

// Handle input and output, processing logic.
func (c *cacheServiceImpl) Handle(in []byte) (out []byte) {
	out = in
	return
}

func (c *cacheServiceImpl) Write(context.Context, *CacheWriter) (*CacheTtl, error) {
	return &CacheTtl{Ttl: 0}, nil
}

func (c *cacheServiceImpl) Read(context.Context, *CacheReader) (*CacheValue, error) {
	return &CacheValue{Value: []byte{}}, nil
}

func (c *cacheServiceImpl) Delete(context.Context, *CacheReader) (*CacheTtl, error) {
	return &CacheTtl{Ttl: 0}, nil
}

func (c *cacheServiceImpl) Ttl(context.Context, *CacheReader) (*CacheTtl, error) {
	return &CacheTtl{Ttl: 0}, nil
}
