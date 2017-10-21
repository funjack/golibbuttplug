package golibbuttplug

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/funjack/golibbuttplug/buttplugtest"
)

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

// TestButtplugClient only tests if there are no errors when talking with a
// (fake) buttplug server.
func TestButtplugClient(t *testing.T) {
	s := buttplugtest.DefaultTestServer
	http.Handle("/", s)
	ts := httptest.NewServer(s)
	defer ts.Close()

	// Contexts can be used to cancel client connection.
	rootctx := context.Background()
	// Create a new session with the server as "ExampleClient".
	c, err := NewClient(rootctx, makeWsProto(ts.URL), "ExampleClient", nil)
	if err != nil {
		log.Fatal(err)
	}
	// Scan for devices.
	if err := c.StartScanning(); err != nil {
		log.Fatal(err)
	}
	// Simulate some events.
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.Conn.SendScanningFinished()
		time.Sleep(10 * time.Millisecond)
		s.Conn.AddDevice(buttplugtest.DefaultAddDeviceMessage)
		time.Sleep(10 * time.Millisecond)
		s.Conn.RemoveDevice(buttplugtest.DefaultAddDeviceMessage)
	}()
	// Wait for scanning to finish.
	ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
	err = c.WaitOnScanning(ctx)
	cancel()
	if err == context.DeadlineExceeded {
		// Stop scanning.
		if err := c.StopScanning(); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}
	// Get all known devices.
	log.Printf("devices: %v", c.Devices())
	for _, d := range c.Devices() {
		go HandleDisconnect(d)
		// Test if the RawCmd is supported by the device.
		if d.IsSupported("RawCmd") {
			log.Printf("%s supports RawCmd", d.Name())
			// Send a RawCmd.
			if err := d.RawCmd([]byte{0x00, 0x10}); err != nil {
				log.Printf("RawCmd failed: %v", err)
			}
		}
		// Try sending a Fleshlight Launch command.
		if err := d.FleshlightLaunchFW12Cmd(50, 20); err != nil {
			log.Printf("Launch command failed: %v", err)
		}
	}
	time.Sleep(200 * time.Millisecond)
	// Stop all devices.
	if err := c.StopAllDevices(); err != nil {
		log.Fatal(err)
	}
	// Close the connection.
	c.Close()
	time.Sleep(100 * time.Millisecond)
}

func HandleDisconnect(d *Device) {
	<-d.Disconnected()
	log.Printf("Lost device: %s", d.Name())
}
