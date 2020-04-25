package consumer

import (
	"github.com/angenalZZZ/gofunc/data/queue/message"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	topic       = "consumerTestTopic"
	channel     = "consumerTestChannel"
	maxInFlight = 3

	lookupDHTTPAddr = "127.0.0.1:4161"
	//lookupDHTTPAdds = []string{"127.0.0.2:4161"}

	destNsqDTCPAddr = "127.0.0.1:4150"
	//destNsqDTCPAdds = []string{"127.0.0.2:4150"}
)

func mgsHandle(_ *message.NsqMessage) error { return nil }

func TestRegister(t *testing.T) {
	Convey("Given topic, channel, maxInflight and message handler method", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.Register(topic, channel, maxInFlight, mgsHandle)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Given wrong topic, channel", t, func() {
		Convey("It should produce an error", func() {
			c := NewNsqConsumer()
			err := c.Register("", "", maxInFlight, mgsHandle)
			So(err, ShouldNotEqual, nil)
		})
	})
}

func TestConnectLookupD(t *testing.T) {
	Convey("Given lookupD address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.ConnectLookupD(lookupDHTTPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Given wrong lookupD address", t, func() {
		Convey("It should produce an error", func() {
			c := NewNsqConsumer()
			err := c.ConnectLookupD("127.0.0.1")
			So(err, ShouldNotEqual, nil)
		})
	})
}

func TestConnect(t *testing.T) {
	Convey("Given nsqD address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.Connect(destNsqDTCPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Given wrong nsqD address", t, func() {
		Convey("It should produce an error", func() {
			c := NewNsqConsumer()
			err := c.Connect("127.0.0.1")
			So(err, ShouldNotEqual, nil)
		})
	})
}
