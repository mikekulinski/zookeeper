package session

import (
	pbzk "github.com/mikekulinski/zookeeper/proto"
)

type Session struct {
	// Messages is a channel of events that the server needs to process.
	Messages chan *Event
}

func NewSession() *Session {
	return &Session{
		// Messages is intentionally not buffered so we can check for timeouts.
		Messages: make(chan *Event),
	}
}

// An Event is a message that the server needs to process. This can be of
// several different types. We only expect one of these fields to be non-nil.
type Event struct {
	ClientRequest *pbzk.ZookeeperRequest
	WatchEvent    *pbzk.WatchEvent
	// EOF is used to tell the server that we have lost connection with the client.
	// We use this instead of closing the channel since we have multiple writers to the channel.
	EOF bool
}
