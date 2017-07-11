/*
Package golibbuttplug provides a Buttplug websocket client.

Buttplug (https://buttplug.io/) is a quasi-standard set of technologies and
protocols to allow developers to write software that controls an array of sex
toys in a semi-future-proof way.
*/
package golibbuttplug

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/funjack/golibbuttplug/message"
	"github.com/gorilla/websocket"
)

// DefaultName is used when no name is specified when creating a new client.
const DefaultName = "golibbuttplug"

var defaultTimeout = time.Second * 30

// Client is a websocket API client that performs operations against a Buttplug
// server.
type Client struct {
	ctx     context.Context
	conn    *websocket.Conn    // Websocket connection with Buttplug server.
	counter *message.IDCounter // Message ID counter

	once     sync.Once         // Ensure Close() is executed only once.
	stop     chan struct{}     // Halts pingLoop and eventLoop goroutines.
	sender   *message.Sender   // Sending messages.
	receiver *message.Receiver // Receiving messages.

	m       sync.RWMutex       // Protects devices map.
	devices map[uint32]*Device // Devices by their DeviceIndex
}

// NewClient returns a new client with a connection to a Buttplug server.
func NewClient(ctx context.Context, addr, name string) (c *Client, err error) {
	c = &Client{
		ctx:     ctx,
		counter: new(message.IDCounter),
		stop:    make(chan struct{}),
		devices: make(map[uint32]*Device),
	}
	// Create websocket connection
	u, err := url.ParseRequestURI(addr)
	if err != nil {
		return nil, err
	}
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	// Start the reader and writer.
	c.receiver = message.NewReceiver(c.conn, c.stop)
	c.sender = message.NewSender(c.conn)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.stop:
		}
		c.Close()
	}()
	// Initialize a session with the server.
	if name == "" {
		name = DefaultName
	}
	err = c.initSession(name)
	if err != nil {
		c.Close()
		return nil, err
	}
	// Setup the initial device list.
	err = c.initDeviceList()
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// Close the connection.
func (c *Client) Close() {
	c.once.Do(func() {
		log.Printf("Closing connection to Buttplug")
		c.sender.Stop()
		c.receiver.Stop()
		<-c.stop
		c.conn.Close()
		log.Printf("Connection to Buttplug closed")
	})
}

// InitSession creates a session with server by requesting serverinfo and
// starting a ping/pong exchange.
func (c *Client) initSession(name string) error {
	// Send RequestServerInfo
	id := c.counter.Generate()
	r := message.OutgoingMessage{
		RequestServerInfo: &message.RequestServerInfo{
			ID:         id,
			ClientName: name,
		},
	}
	if err := c.sender.Send(r); err != nil {
		return err
	}
	// Read reply
	ctx, cancel := context.WithTimeout(c.ctx, defaultTimeout)
	defer cancel()
	m, err := c.receiveMessage(ctx, id)
	if err != nil {
		return err
	}
	if m.ServerInfo == nil {
		return errors.New("no serverinfo received")
	}
	si := *m.ServerInfo
	log.Printf("Connected to Buttplug %s (%d.%d.%d)", si.ServerName,
		si.BuildVersion, si.MajorVersion, si.MinorVersion)
	// Start ping goroutine
	interval := time.Duration(1) * time.Second
	if si.MaxPingTime != 0 && si.MaxPingTime < 1000 {
		interval = time.Duration(si.MaxPingTime/2) * time.Millisecond
	}
	go c.pingLoop(interval)
	return nil
}

// PingLoop sends out pings.
func (c *Client) pingLoop(d time.Duration) {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.stop:
			return
		case <-time.After(d):
			id := c.counter.Generate()
			m := message.OutgoingMessage{
				Ping: &message.Empty{
					ID: id,
				},
			}
			if err := c.sendMessage(id, m); err != nil {
				log.Printf("ping error: %v", err)
				c.Close()
			}
		}
	}
}

// InitDeviceList syncs up client device list with server.
func (c *Client) initDeviceList() error {
	// Send RequestDeviceList
	id := c.counter.Generate()
	r := message.OutgoingMessage{
		RequestDeviceList: &message.Empty{
			ID: id,
		},
	}
	if err := c.sender.Send(r); err != nil {
		return err
	}
	// Retreive response
	ctx, cancel := context.WithTimeout(c.ctx, defaultTimeout)
	defer cancel()
	m, err := c.receiveMessage(ctx, id)
	if err != nil {
		return err
	}
	// Update DeviceList
	dl := *m.DeviceList
	for _, d := range dl.Devices {
		c.addDevice(d)
	}
	// Start event watcher goroutine.
	s := c.receiver.Subscribe()
	go c.eventLoop(s)
	return nil
}

// EventLoop watches for (device) events.
func (c *Client) eventLoop(in *message.Reader) {
	for m := range in.Incoming() {
		if m.DeviceAdded != nil {
			c.addDevice(*m.DeviceAdded)
		}
		if m.DeviceRemoved != nil {
			c.removeDevice(*m.DeviceRemoved)
		}
	}
}

// AddDevice to the device list.
func (c *Client) addDevice(d message.Device) {
	c.m.Lock()
	defer c.m.Unlock()
	log.Printf("Found device: %s (%d)", d.DeviceName, d.DeviceIndex)
	c.devices[d.DeviceIndex] = &Device{
		client: c,
		device: d,
		done:   make(chan struct{}),
	}
}

// RemoveDevice from the device list.
func (c *Client) removeDevice(d message.Device) {
	c.m.Lock()
	defer c.m.Unlock()
	log.Printf("Removed device: %s (%d)",
		c.devices[d.DeviceIndex].device.DeviceName, d.DeviceIndex)
	if dev, ok := c.devices[d.DeviceIndex]; ok {
		close(dev.done)
	}
	delete(c.devices, d.DeviceIndex)
}

// ReceiveMessage waits for and reads a message with a given id.
func (c *Client) receiveMessage(ctx context.Context, id uint32) (message.IncomingMessage, error) {
	r := c.receiver.Subscribe()
	defer c.receiver.Unsubscribe(r)
	for {
		select {
		case msg, ok := <-r.Incoming():
			if !ok {
				return msg, errors.New("reader stopped")
			}
			if msgid, _ := msg.Message(); msgid == id {
				return msg, nil
			}
		case <-ctx.Done():
			return message.IncomingMessage{}, ctx.Err()
		}
	}
}

// SendMessage is a generic send and read Ok/Error message with the default
// timeout.
func (c *Client) sendMessage(id uint32, m message.OutgoingMessage) error {
	if err := c.sender.Send(m); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.ctx, defaultTimeout)
	defer cancel()
	r, err := c.receiveMessage(ctx, id)
	if err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("server error: %s", r.Error.ErrorMessage)
	}
	if r.Ok == nil {
		return errors.New("did not receive ok")
	}
	return nil
}

// StartScanning requests to have the server start scanning for devices on all
// busses that it knows about. Useful for protocols like Bluetooth, which
// require an explicit discovery phase.
func (c *Client) StartScanning() error {
	id := c.counter.Generate()
	m := message.OutgoingMessage{
		StartScanning: &message.Empty{
			ID: id,
		},
	}
	if err := c.sendMessage(id, m); err != nil {
		return err
	}
	return nil
}

// StopScanning requests to have the server stop scanning for devices. Useful
// for protocols like Bluetooth, which may not timeout otherwise.
func (c *Client) StopScanning() error {
	id := c.counter.Generate()
	m := message.OutgoingMessage{
		StopScanning: &message.Empty{
			ID: id,
		},
	}
	return c.sendMessage(id, m)
}

// WaitOnScanning waits until the server has stopped scanning on all busses.
func (c *Client) WaitOnScanning(ctx context.Context) error {
	r := c.receiver.Subscribe()
	defer c.receiver.Unsubscribe(r)
	for {
		select {
		case msg, ok := <-r.Incoming():
			if !ok {
				return errors.New("reader stopped")
			}
			if msg.ScanningFinished != nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Devices returns all devices currently known by the client.
func (c *Client) Devices() []*Device {
	c.m.RLock()
	defer c.m.RUnlock()
	d := make([]*Device, 0, len(c.devices))
	for _, v := range c.devices {
		d = append(d, v)
	}
	return d
}

// StopAllDevices tells the server to stop all devices. Can be used for
// emergency situations, on client shutdown for cleanup, etc.
func (c *Client) StopAllDevices() error {
	id := c.counter.Generate()
	m := message.OutgoingMessage{
		StopAllDevices: &message.Empty{
			ID: id,
		},
	}
	return c.sendMessage(id, m)
}
