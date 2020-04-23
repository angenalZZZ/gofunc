package message

import (
	"context"
	json "github.com/json-iterator/go"
	"github.com/nsqio/go-nsq"
)

type TMessage struct {
	Topic   string
	Channel string
}

// NsqMessage Inherent nsq.
type NsqMessage struct {
	*nsq.Message
}

var msgKey = "nsqMsg"

// WithContext Returns nsq message from context.
func WithContext(ctx context.Context, msg *NsqMessage) context.Context {
	return context.WithValue(ctx, msgKey, msg)
}

// FromContext Returns nsq message from context.
func FromContext(ctx context.Context) (*NsqMessage, bool) {
	value, ok := ctx.Value(msgKey).(*NsqMessage)
	return value, ok
}

// GiveUp Finish message with success state because message never will be possible to process.
func (m *NsqMessage) GiveUp() {
	m.Finish(true)
}

// Success Finish message successfully.
func (m *NsqMessage) Success() {
	m.Finish(true)
}

// Fail Mark message as failed to process.
func (m *NsqMessage) Fail() {
	m.Finish(false)
}

// Finish Processing message.
func (m *NsqMessage) Finish(success bool) {
	if success {
		m.Message.Finish()
	} else {
		m.Message.Requeue(-1)
	}
}

// ReadJSON UnMarshals JSON message body to interface.
func (m *NsqMessage) ReadJSON(v interface{}) error {
	return json.ConfigCompatibleWithStandardLibrary.Unmarshal(m.Body, v)
}
