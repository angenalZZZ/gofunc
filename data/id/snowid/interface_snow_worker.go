package snowid

// ISnowWorker .
type ISnowWorker interface {
	NextId() uint64
}
