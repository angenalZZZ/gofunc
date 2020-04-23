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

	lookupdHTTPAddr = "127.0.0.1:4161"
	//lookupdHTTPAddrs = []string{"127.0.0.2:4161"}

	destNsqdTCPAddr = "127.0.0.1:4150"
	//destNsqdTCPAddrs = []string{"127.0.0.2:4150"}
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

func TestConnectLookupd(t *testing.T) {
	Convey("Given lookupd address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.ConnectLookupD(lookupdHTTPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Given wrong lookupd address", t, func() {
		Convey("It should produce an error", func() {
			c := NewNsqConsumer()
			err := c.ConnectLookupD("127.0.0.1")
			So(err, ShouldNotEqual, nil)
		})
	})
}

func TestConnect(t *testing.T) {
	Convey("Given nsqd address", t, func() {
		Convey("It should not produce any error", func() {
			c := NewNsqConsumer()
			err := c.Connect(destNsqdTCPAddr)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Given wrong nsqd address", t, func() {
		Convey("It should produce an error", func() {
			c := NewNsqConsumer()
			err := c.Connect("127.0.0.1")
			So(err, ShouldNotEqual, nil)
		})
	})
}
