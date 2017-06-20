package client

import (
	"context"
	"log"
	"time"
)

// This example demonstrates how to connect to Buttplug server websocket,
// search for devices and perform operations on the discovered devices.
func ExampleClient() {
	// Contexts can be used to cancel client connection.
	rootctx := context.Background()
	// Create a new session with the server as "ExampleClient".
	c, err := NewClient(rootctx, "ws://127.0.0.1:12345", "ExampleClient")
	if err != nil {
		log.Fatal(err)
	}
	// Scan for devices.
	if err := c.StartScanning(); err != nil {
		log.Fatal(err)
	}
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
	for _, d := range c.Devices() {
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
	// Stop all devices.
	if err := c.StopAllDevices(); err != nil {
		log.Fatal(err)
	}
	// Close the connection.
	c.Close()
}
