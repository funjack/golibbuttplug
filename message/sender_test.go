package message

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func TestIDCounter(t *testing.T) {
	id := IDCounter{}
	if id.Generate() <= 0 {
		t.Errorf("invalid id")
	}
	id.value = 4294967295
	if id.Generate() == 0 {
		t.Errorf("id zero should not be generated")
	}
}

func TestSend(t *testing.T) {
	testSend(t, 1)
}

func testSend(tb testing.TB, n int) {
	start := make(chan struct{})
	done := make(chan struct{})
	var upgrader = websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ws.Close()
			tb.Error(err)
		}
		go readLoop(ws)
		sender := NewSender(ws)

		<-start
		for i := 0; i < n; i++ {
			sender.Send(OutgoingMessage{
				Ping: &Empty{ID: uint32(i)},
			})
		}
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			tb.Errorf("test timeout")
		}
		sender.Stop()
		select {
		case <-sender.stop:
		case <-time.After(10 * time.Second):
			tb.Errorf("sender was not stopped withing timeout")
		}
	}))
	defer s.Close()

	conn, _, err := websocket.DefaultDialer.Dial(makeWsProto(s.URL), nil)
	if err != nil {
		tb.Error(err)
	}
	defer conn.Close()
	close(start)
	for {
		var msgs []OutgoingMessage
		err := conn.ReadJSON(&msgs)
		if err != nil {
			tb.Error(err)
			continue
		}
		if len(msgs) < 1 {
			tb.Errorf("empty message list received")

		}
		if msgs[0].Ping == nil {
			tb.Errorf("ping message not received")
		}
		if msgs[0].Ping.ID >= uint32(n-1) {
			close(done)
			return
		}
	}
}

func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			return
		}
	}
}
