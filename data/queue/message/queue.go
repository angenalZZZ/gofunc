package message

import (
	"github.com/nsqio/go-nsq"
)

type Queue struct {
	H Handler
	*nsq.Consumer
}

func (q *Queue) HandleMessage(message *nsq.Message) error {
	return q.H(&NsqMessage{message})
}
