package producer

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	topic           = "producerTestTopic"
	destNSQdTCPAddr = "127.0.0.1:4150"
)

func TestNSQPublish(t *testing.T) {
	Convey("Given a json message to publish", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message = []byte{0x18}
			err = p.Publish(topic, message)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQPublishAsync(t *testing.T) {
	Convey("Given a json message to publish asynchronously", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message = []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
			err = p.PublishAsync(topic, message, nil)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQMultiPublish(t *testing.T) {
	Convey("Given a multiple message to publish", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message1 = []byte{0x18}
			var message = [][]byte{message1}
			err = p.MultiPublish(topic, message)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQMultiPublishAsync(t *testing.T) {
	Convey("Given a multiple message to publish asynchrnously", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message1 = []byte{0x18}
			var message = [][]byte{message1}
			err = p.MultiPublishAsync(topic, message, nil)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQPublishJSONAsync(t *testing.T) {
	Convey("Given a topic and a message to publish asynchronously", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message interface{} = "testMessage"
			err = p.PublishJSONAsync(topic, message, nil)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQPublishJSON(t *testing.T) {
	Convey("Given topic to publish a json message", t, func() {
		Convey("It should not produce any error", func() {
			p, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
			var message interface{} = "testMessage"
			err = p.PublishJSON(topic, message)
			So(err, ShouldEqual, nil)
		})
	})
}

func TestNSQConnect(t *testing.T) {
	Convey("Given nsqd address to connect to", t, func() {
		Convey("It should not produce any error", func() {
			_, err := NewNsqProducer(destNSQdTCPAddr)
			So(err, ShouldEqual, nil)
		})
	})
}
