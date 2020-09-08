package f

import "sync"

type channelMessage struct {
	v  interface{}
	ok bool
}

// C is a channel
type Channel interface {
	// Send a messge to the channel. Returns false if the channel is closed.
	Send(v interface{}) (ok bool)
	// Recv a messge from the channel. Returns false if the channel is closed.
	Recv() (v interface{}, ok bool)
	// Close the channel. Returns false if the channel is already closed.
	Close() (ok bool)
	// Wait for the channel to close. Returns immediately if the channel is
	// already closed
	Wait()
}

type channel struct {
	mu     sync.Mutex
	cond   *sync.Cond
	c      chan channelMessage
	closed bool
}

// Make new channel. Provide a length to make a buffered channel.
func MakeChannel(length int) Channel {
	c := &channel{c: make(chan channelMessage, length)}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *channel) Send(v interface{}) (ok bool) {
	defer func() { ok = recover() == nil }()
	c.c <- channelMessage{v, true}
	return
}

func (c *channel) Recv() (v interface{}, ok bool) {
	select {
	case msg := <-c.c:
		return msg.v, msg.ok
	}
}

func (c *channel) Close() (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() { ok = recover() == nil }()
	close(c.c)
	c.closed = true
	c.cond.Broadcast()
	return
}

func (c *channel) Wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		if c.closed {
			return
		}
		c.cond.Wait()
	}
}
