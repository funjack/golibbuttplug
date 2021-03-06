package message

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// Receiver can read Buttplug server messages from a websocket to multiple
// readers. Readers can subscribe/unsubscribe from receiving messages.
type Receiver struct {
	once sync.Once // Make sure Stop() is execute only once.
	conn *websocket.Conn
	hub  *hub
}

// NewReceiver creates a Receiver for the given websocket connection. Done
// channel is closed then receiver is done.
func NewReceiver(conn *websocket.Conn, done chan struct{}) *Receiver {
	r := &Receiver{
		conn: conn,
		hub:  newHub(),
	}
	go r.run(done)
	return r
}

// Run reads a message from the websocket connection and puts it on the hub.
func (rc *Receiver) run(done chan struct{}) {
	for {
		var msgs IncomingMessages
		err := rc.conn.ReadJSON(&msgs)
		if err != nil {
			rc.conn.Close()
			close(done)
			return
		}
		for _, msg := range msgs {
			select {
			case rc.hub.incoming <- msg:
			case <-rc.hub.stop:
			}
		}
	}

}

// Subscribe creates a new reader that receives messages. A consumer should
// call the Unsubscribe when it's done with the reader.
func (rc *Receiver) Subscribe() (*Reader, error) {
	r := &Reader{
		buf: make(chan IncomingMessage, 10),
	}
	select {
	case rc.hub.subscribe <- r:
		return r, nil
	case <-rc.hub.stop:
		return nil, errors.New("stopped")
	}
}

// Unsubscribe removes the readers subscription and will no longer receive
// messages.
func (rc *Receiver) Unsubscribe(r *Reader) {
	select {
	case rc.hub.unsubscribe <- r:
	case <-rc.hub.stop:
	}
}

// Stop the receiver from sending any messages.
func (rc *Receiver) Stop() {
	rc.once.Do(func() {
		close(rc.hub.stop)
	})
}

// Reader is receives messages from the Receiver subscription.
type Reader struct {
	// buffered channel for subscriber to read
	buf chan IncomingMessage
}

// Incoming returns a channel of incoming messages.
func (r *Reader) Incoming() <-chan IncomingMessage {
	return r.buf
}

// Hub forwards messages to subscribed readers.
type hub struct {
	readers map[*Reader]bool
	// incoming message
	incoming chan IncomingMessage
	// subscribe a reader to receive messages
	subscribe chan *Reader
	// unsubscribe a reader from receiving messages
	unsubscribe chan *Reader
	// stop the hub
	stop chan bool
}

// Hub broadcasts messages received on the incoming channel to all subscribed
// readers.
func newHub() *hub {
	r := &hub{
		readers:     make(map[*Reader]bool),
		incoming:    make(chan IncomingMessage),
		subscribe:   make(chan *Reader),
		unsubscribe: make(chan *Reader),
		stop:        make(chan bool),
	}
	go r.run()
	return r
}

func (h *hub) run() {
	for {
		select {
		case <-h.stop:
			for reader := range h.readers {
				close(reader.buf)
				delete(h.readers, reader)
			}
			return
		case reader := <-h.subscribe:
			h.readers[reader] = true
		case reader := <-h.unsubscribe:
			if _, ok := h.readers[reader]; ok {
				close(reader.buf)
				delete(h.readers, reader)
			}
		case msg := <-h.incoming:
			for reader := range h.readers {
				select {
				case reader.buf <- msg:
				default:
					close(reader.buf)
					delete(h.readers, reader)
				}
			}
		}
	}
}
