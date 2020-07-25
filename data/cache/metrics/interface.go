package metrics

import "github.com/angenalZZZ/gofunc/data/cache/codec"

// MetricsInterface represents the metrics interface for all available providers
type MetricsInterface interface {
	RecordFromCodec(codec codec.CodecInterface)
}
