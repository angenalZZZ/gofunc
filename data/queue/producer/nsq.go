package producer

import (
	"github.com/angenalZZZ/gofunc/log"
	"github.com/nsqio/go-nsq"
)

type NsqProducer struct {
	nsq.LogLevel
	*nsq.Producer
}

func NewNsqProducer() *NsqProducer {
	return &NsqProducer{
		LogLevel: log.Log.GetLevel(),
		Producer: nil,
	}
}
