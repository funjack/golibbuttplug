// Package buttplugtest provides utilities for buttplug testing.
package buttplugtest

import (
	"log"
	"net/http"
	"sync"

	"github.com/funjack/golibbuttplug/message"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// TestServer is a mock of a Buttplug server.
type TestServer struct {
	InitialDevices []message.Device
	Conn           *Conn
}

var (
	// DefaultAddDeviceMessage can be used to simulate adding a Launch.
	DefaultAddDeviceMessage = &message.Device{
		ID:             0,
		DeviceName:     "Launch",
		DeviceIndex:    3,
		DeviceMessages: []string{"FleshlightLaunchFW12Cmd", "KiirooCmd", "RawCmd", "StopDeviceCmd"},
	}
	// DefaultRemoveDeviceMessage can be used to remove the added Launch.
	DefaultRemoveDeviceMessage = &message.Device{
		ID:          0,
		DeviceIndex: 3,
	}
)

// DefaultTestServer is a TestServer with some predefined devices.
var DefaultTestServer = &TestServer{
	InitialDevices: []message.Device{
		{
			DeviceName:     "TestDevice 1",
			DeviceIndex:    0,
			DeviceMessages: []string{"SingleMotorVibrateCmd", "RawCmd", "KiirooCmd", "StopDeviceCmd"},
		},
		{
			DeviceName:     "TestDevice 2",
			DeviceIndex:    1,
			DeviceMessages: []string{"SingleMotorVibrateCmd", "LovenseCmd", "StopDeviceCmd"},
		},
		{
			DeviceName:     "Launch",
			DeviceIndex:    2,
			DeviceMessages: []string{"FleshlightLaunchFW12Cmd", "KiirooCmd", "RawCmd", "StopDeviceCmd"},
		},
	},
}

func (t *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()
	t.Conn = &Conn{
		conn:    conn,
		devices: t.InitialDevices,
	}
	err = t.Conn.ReadMessages()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	return
}

// Conn is an established websocket connection with the testserver.
type Conn struct {
	sync.Mutex
	conn    *websocket.Conn
	devices []message.Device
}

// ReadMessages will read the messages from the websocket to be read and handled.
func (c *Conn) ReadMessages() error {
	for {
		var msgs message.OutgoingMessages
		err := c.conn.ReadJSON(&msgs)
		if _, ok := err.(*websocket.CloseError); ok {
			return err
		} else if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}
		for _, msg := range msgs {
			c.handleMessage(msg)
		}
	}
}

func (c *Conn) handleMessage(m message.OutgoingMessage) {
	switch true {
	case m.RequestServerInfo != nil:
		id := m.RequestServerInfo.ID
		log.Printf("<-RequestServerInfo (%d)", id)
		c.sendServerInfo(id)
	case m.RequestDeviceList != nil:
		id := m.RequestDeviceList.ID
		log.Printf("<-RequestDeviceList (%d)", id)
		c.sendDeviceList(id)
	case m.StopScanning != nil:
		id := m.StopScanning.ID
		log.Printf("<-StopScanning (%d)", id)
		c.sendOk(id)
	case m.Ping != nil:
		id := m.Ping.ID
		log.Printf("<-Ping (%d)", id)
		c.sendOk(id)
	case m.FleshlightLaunchFW12Cmd != nil:
		id := m.FleshlightLaunchFW12Cmd.ID
		pos, spd := m.FleshlightLaunchFW12Cmd.Position, m.FleshlightLaunchFW12Cmd.Speed
		log.Printf("<-FleshlightLaunchFW12Cmd (%d) Postion = %d, Speed = %d", id, pos, spd)
		c.sendOk(id)
	case m.KiirooCmd != nil:
		id := m.KiirooCmd.ID
		log.Printf("<-KiirooCmd (%d)", id)
		c.sendOk(id)
	case m.LovenseCmd != nil:
		id := m.LovenseCmd.ID
		log.Printf("<-LovenseCmd (%d)", id)
		c.sendOk(id)
	case m.VorzeA10CycloneCmd != nil:
		id := m.VorzeA10CycloneCmd.ID
		spd := m.VorzeA10CycloneCmd.Speed
		clockwise := m.VorzeA10CycloneCmd.Clockwise
		log.Printf("<-VorzeA10CycloneCmd (%d) Speed = %d, Clockwise: = %t", id, spd, clockwise)
		c.sendOk(id)
	case m.RawCmd != nil:
		id := m.RawCmd.ID
		log.Printf("<-RawCmd (%d)", id)
		c.sendOk(id)
	case m.StartScanning != nil:
		id := m.StartScanning.ID
		log.Printf("<-StartScanning, (%d)", id)
		c.sendOk(id)
	case m.StopAllDevices != nil:
		id := m.StopAllDevices.ID
		log.Printf("<-StopAllDevices (%d)", id)
		c.sendOk(id)
	case m.StopDeviceCmd != nil:
		id := m.StopDeviceCmd.ID
		log.Printf("<-StopDeviceCmd (%d)", id)
		c.sendOk(id)
	}
}

func (c *Conn) sendOk(id uint32) {
	msg := message.IncomingMessage{
		Ok: &message.Empty{
			ID: id,
		},
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Printf("->Ok (%d)", id)
}

func (c *Conn) sendServerInfo(id uint32) {
	msg := message.IncomingMessage{
		ServerInfo: &message.ServerInfo{
			ID:             id,
			ServerName:     "TestButtplug",
			MessageVersion: 1,
			MajorVersion:   1,
			MinorVersion:   0,
			BuildVersion:   0,
			MaxPingTime:    100,
		},
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Printf("->ServerInfo (%d)", id)
}

func (c *Conn) sendDeviceList(id uint32) {
	msg := message.IncomingMessage{
		DeviceList: &message.DeviceList{
			ID:      id,
			Devices: c.devices,
		},
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Printf("->DeviceList (%d)", id)
}

// SendScanningFinished will send a message to the client that scanning is
// finished.
func (c *Conn) SendScanningFinished() {
	msg := message.IncomingMessage{
		ScanningFinished: &message.Empty{
			ID: 0,
		},
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Println("->ScanningFinished (0)")
}

// AddDevice will send the a message to the client that the given device has
// been added.
func (c *Conn) AddDevice(d *message.Device) {
	msg := message.IncomingMessage{
		DeviceAdded: d,
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Println("->DeviceAdded (0)")
}

// RemoveDevice will send the a message to the client that the given device has
// been removed.
func (c *Conn) RemoveDevice(d *message.Device) {
	msg := message.IncomingMessage{
		DeviceRemoved: d,
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteJSON(message.IncomingMessages{msg})
	if err != nil {
		log.Printf("error writing: %v", err)
	}
	log.Println("->DeviceRemoved (0)")
}
