package message

// Handler function that handles incoming message.
type Handler func(*NsqMessage) error
