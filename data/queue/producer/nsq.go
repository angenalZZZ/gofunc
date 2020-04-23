package producer

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/log"
	json "github.com/json-iterator/go"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog"
	"runtime"
)

type NsqProducer struct {
	Log *log.Logger
	*nsq.Producer
}

// NewNsqProducer create a nsq producer.
func NewNsqProducer(addr string, config ...*nsq.Config) (p *NsqProducer, err error) {
	var (
		l  *log.Logger
		ll zerolog.Level
		np *nsq.Producer
		cf *nsq.Config
	)

	if log.Log != nil {
		l = log.Log
	} else {
		l = log.InitConsole("2006-01-02 15:04:05.000", false)
		ll = zerolog.ErrorLevel
		l.Level(ll)
	}

	if len(config) > 0 {
		cf = config[0]
	} else {
		cf = nsq.NewConfig()
	}

	np, err = nsq.NewProducer(addr, cf)
	if err != nil {
		return
	}

	np.SetLogger(p, nsq.LogLevel(int(ll)))
	p = &NsqProducer{
		Log:      l,
		Producer: np,
	}
	return
}

// PublishJSONAsync sends message to nsq  topic in json format asynchronously.
func (p *NsqProducer) PublishJSONAsync(topic string, v interface{}, doneChan chan *nsq.ProducerTransaction,
	args ...interface{}) error {
	body, err := json.ConfigCompatibleWithStandardLibrary.Marshal(v)
	if err != nil {
		return err
	}
	return p.Producer.PublishAsync(topic, body, doneChan, args...)
}

// PublishJSON sends message to nsq  topic in json format.
func (p *NsqProducer) PublishJSON(topic string, v interface{}) error {
	body, err := json.ConfigCompatibleWithStandardLibrary.Marshal(v)
	if err != nil {
		return err
	}
	return p.Producer.Publish(topic, body)
}

// Output log.
func (p *NsqProducer) Output(calldepth int, s string) error {
	if p.Log.GetLevel() == zerolog.DebugLevel {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		s = fmt.Sprintf("%s %04d: %s", file, line, s)
	}
	p.Log.Print(s)
	return nil
}
