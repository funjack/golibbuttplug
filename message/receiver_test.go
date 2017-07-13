package message

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestReceive(t *testing.T) {
	testReceive(t, 10, 1)
}

func TestReceiveMultipleSubs(t *testing.T) {
	testReceive(t, 10, 5)
}

func testReceive(tb testing.TB, nMsg, nSubs int) {
	start := make(chan struct{}) // start channel start mock server sending
	done := make(chan struct{})  // done stops mockserver

	var upgrader = websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ws.Close()
			tb.Error(err)
		}
		go readLoop(ws)
		<-start
		for i := 0; true; i++ {
			err = ws.WriteJSON(IncomingMessages{
				{
					Ok: &Empty{
						ID: uint32(i),
					},
				},
			})
		}
		<-done
	}))
	defer s.Close()

	conn, _, err := websocket.DefaultDialer.Dial(makeWsProto(s.URL), nil)
	if err != nil {
		tb.Error(err)
	}
	defer conn.Close()

	stopchan := make(chan struct{})
	receiver := NewReceiver(conn, stopchan)

	var wg sync.WaitGroup
	for i := 0; i < nSubs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := receiver.Subscribe()
			if err != nil {
				return
			}
			defer receiver.Unsubscribe(r)
			for {
				select {
				case m, ok := <-r.Incoming():
					if !ok {
						tb.Error("incoming channel closed before receiving all messages")
						return
					}
					if m.Ok == nil {
						tb.Errorf("no ok mesg received %+v", m)
						return
					}
					if m.Ok.ID >= uint32(nMsg-1) {
						return
					}
				}
			}
		}()
	}
	close(start) // start sending messages from the mock server
	wg.Wait()    // wait until all subs are done
	receiver.Stop()

	// Check if everything is done and stopped
	select {
	case <-receiver.hub.stop:
	case <-time.After(10 * time.Second):
		tb.Errorf("receiver not stopped")
	}

	close(done)  // stop mock server
	conn.Close() // close websocket

	select {
	case <-stopchan:
	case <-time.After(10 * time.Second):
		tb.Errorf("receiver didn't close stopchan")
	}
}
