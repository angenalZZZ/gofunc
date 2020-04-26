package consumer

import (
	"github.com/angenalZZZ/gofunc/data/queue/message"
	"github.com/angenalZZZ/gofunc/data/queue/nsq"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func mgsHandle(m *message.NsqMessage) error {
	println(m.Message)
	return nil // handle message finish
}

func TestRegister(t *testing.T) {
	Convey("Given topic, channel, maxInflight and message handler method", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.Register(nsq.TestTopic, nsq.TestChannel, 2, mgsHandle)
			So(err, ShouldEqual, nil)
			time.Sleep(time.Second)
			c.Stop()
		})
	})

	//Convey("Given wrong topic, channel", t, func() {
	//	Convey("It should produce an error", func() {
	//		c := NewNsqConsumer()
	//		err := c.Register("", "", maxInFlight, mgsHandle)
	//		So(err, ShouldNotEqual, nil)
	//	})
	//})
}

func TestConnectLookupD(t *testing.T) {
	Convey("Given lookupD address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.ConnectLookupD(nsq.LOOKUPdHTTPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	//Convey("Given wrong lookupD address", t, func() {
	//	Convey("It should produce an error", func() {
	//		c := NewNsqConsumer()
	//		err := c.ConnectLookupD("127.0.0.1")
	//		So(err, ShouldNotEqual, nil)
	//	})
	//})
}

func TestConnect(t *testing.T) {
	Convey("Given nsqD address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.Connect(nsq.NSQdTCPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	//Convey("Given wrong nsqD address", t, func() {
	//	Convey("It should produce an error", func() {
	//		c := NewNsqConsumer()
	//		err := c.Connect("127.0.0.1")
	//		So(err, ShouldNotEqual, nil)
	//	})
	//})
}
