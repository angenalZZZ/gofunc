// +build darwin netbsd freebsd openbsd dragonfly

package netpoll

import "golang.org/x/sys/unix"

const (
	// InitEvents represents the initial length of poller event-list.
	InitEvents = 64
	// EVFilterWrite represents writeable events from sockets.
	EVFilterWrite = unix.EVFILT_WRITE
	// EVFilterRead represents readable events from sockets.
	EVFilterRead = unix.EVFILT_READ
	// EVFilterSock represents exceptional events that are not read/write, like socket being closed,
	// reading/writing from/to a closed socket, etc.
	EVFilterSock = -0xd
)

type eventList struct {
	size   int
	events []unix.Kevent_t
}

func newEventList(size int) *eventList {
	return &eventList{size, make([]unix.Kevent_t, size)}
}

func (el *eventList) increase() {
	el.size <<= 1
	el.events = make([]unix.Kevent_t, el.size)
}
