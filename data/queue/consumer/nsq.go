package consumer

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/queue/message"
	"github.com/angenalZZZ/gofunc/log"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"runtime"
)

// NsqConsumer NSQ messages consumer.
type NsqConsumer struct {
	C        chan struct{}
	Log      *log.Logger
	Config   *nsq.Config
	handlers map[message.TMessage]*message.Queue
}

func NewNsqConsumer() *NsqConsumer {
	var (
		l  *log.Logger
		ll zerolog.Level
	)

	if log.Log != nil {
		l = log.Log
	} else {
		l = log.InitConsole("2006-01-02 15:04:05.000", false)
		ll = zerolog.ErrorLevel
		l.Level(ll)
	}

	return &NsqConsumer{
		C:        make(chan struct{}),
		Log:      l,
		Config:   nil,
		handlers: make(map[message.TMessage]*message.Queue),
	}
}

// Register create topic/channel handler for messages,
// This function creates a new nsq Reader.
func (c *NsqConsumer) Register(topic, channel string, maxInFlight int, handler message.Handler) error {
	tch := message.TMessage{Topic: topic, Channel: channel}

	var config *nsq.Config
	if c.Config == nil {
		config = nsq.NewConfig()
		_ = config.Set("verbose", false)
		_ = config.Set("max_in_flight", maxInFlight)
	} else {
		config = c.Config
	}

	r, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	r.SetLogger(c, nsq.LogLevel(int(c.Log.GetLevel())))

	q := &message.Queue{H: handler, Consumer: r}
	r.AddConcurrentHandlers(q, maxInFlight)
	c.handlers[tch] = q
	return nil
}

// Connect - Connects all readers to NSQ.
func (c *NsqConsumer) Connect(addr ...string) error {
	for _, q := range c.handlers {
		for _, add := range addr {
			if err := q.ConnectToNSQD(add); err != nil {
				return err
			}
		}
	}
	return nil
}

// ConnectLookupD Connects all readers to NSQ LookupD.
func (c *NsqConsumer) ConnectLookupD(addr ...string) error {
	for _, q := range c.handlers {
		for _, add := range addr {
			if err := q.ConnectToNSQLookupd(add); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start Just waits.
func (c *NsqConsumer) Start() error {
	if c.C == nil {
		c.C = make(chan struct{})
	}
	<-c.C
	return nil
}

// Stop Gracefully closes all consumers.
func (c *NsqConsumer) Stop() {
	for _, h := range c.handlers {
		h.Stop()
	}
	if c.C != nil {
		close(c.C)
	}
}

// Output log.
func (c *NsqConsumer) Output(calldepth int, s string) error {
	if c.Log.GetLevel() == zerolog.DebugLevel {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		s = fmt.Sprintf("%s %04d: %s", file, line, s)
	}
	c.Log.Print(s)
	return nil
}
