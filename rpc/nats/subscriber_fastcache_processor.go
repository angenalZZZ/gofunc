package nats

// CachePoolWorker is a fast cache Worker.
type CachePoolWorker struct {
	processor func(msg *CacheMsg) error
}

// CacheMsg is a structure used by Subscribers.
type CacheMsg struct {
	Key uint64
	Val []byte
}

func (w *CachePoolWorker) Process(payload interface{}) interface{} {
	return w.processor(payload.(*CacheMsg))
}

func (w *CachePoolWorker) BlockUntilReady() {}
func (w *CachePoolWorker) Interrupt()       {}
func (w *CachePoolWorker) Terminate()       {}
