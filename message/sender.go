package message

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// buffersize is the amount of messages buffered by the Sender.
const bufferSize = 256

// IDCounter is a concurrent safe id counter used for message id's.
type IDCounter struct {
	sync.Mutex
	value uint32
}

// Generate creates a new ID.
func (c *IDCounter) Generate() uint32 {
	c.Lock()
	defer c.Unlock()
	c.value++
	if c.value == 0 {
		c.value = 1
	}
	return c.value
}

// Sender buffers and sends Buttplug messages over a websocket connection.
type Sender struct {
	out  chan<- OutgoingMessage // buffered channelfor outgoing messages.
	once sync.Once              // Make sure Stop() is execute only once.
	stop chan bool
}

// NewSender creates a Sender for the given websocket.
func NewSender(conn *websocket.Conn) (b *Sender) {
	out := make(chan OutgoingMessage, bufferSize)
	b = &Sender{
		stop: make(chan bool),
		out:  out,
	}
	go writeLoop(conn, out)
	return
}

// writeLoop reads messages from buffer and sends them over the websocket.
func writeLoop(conn *websocket.Conn, buf <-chan OutgoingMessage) {
	for v := range buf {
		err := conn.WriteJSON(OutgoingMessages{v})
		if err == websocket.ErrCloseSent {
			return
		} else if err != nil {
			log.Printf("error during write: %v", err)
		}
	}
	err := conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		log.Println("error closing websocket:", err)
	}
}

// Send a message to the server.
func (b *Sender) Send(m OutgoingMessage) error {
	select {
	case <-b.stop:
		return errors.New("stopped")
	case b.out <- m:
		return nil
	default:
		return errors.New("write buffer full")
	}
}

// Stop causes the sender to stop sending messages.
func (b *Sender) Stop() {
	b.once.Do(func() {
		close(b.stop)
		close(b.out)
	})
}
