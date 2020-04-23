package message

import "github.com/nsqio/go-nsq"

// Handler function that handles incoming message.
type Handler func(*nsq.Message) error
