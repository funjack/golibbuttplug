package golibbuttplug

import (
	"errors"
	"fmt"

	"github.com/funjack/golibbuttplug/message"
)

const (
	// CommandStopDevice ...
	CommandStopDevice = "StopDeviceCmd"
	// CommandRaw ...
	CommandRaw = "RawCmd"
	// CommandSingleMotorVibrate ...
	CommandSingleMotorVibrate = "SingleMotorVibrateCmd"
	// CommandKiiroo ...
	CommandKiiroo = "KiirooCmd"
	// CommandFleshlightLaunchFW12 ...
	CommandFleshlightLaunchFW12 = "FleshlightLaunchFW12Cmd"
	// CommandLovense ...
	CommandLovense = "LovenseCmd"
)

var (
	// ErrUnsupported is the error returned when the command executed is
	// not supported by the device.
	ErrUnsupported = errors.New("unsupported command")
	// ErrInvalidSpeed is the error retured when the speed is not supported
	// by the device.
	ErrInvalidSpeed = errors.New("invalid speed")
	// ErrInvalidPosition is the error retured when the position is not
	// supported by the device.
	ErrInvalidPosition = errors.New("invalid position")
	// ErrInvalidCmd is the error retured when the command is not
	// supported by the device.
	ErrInvalidCmd = errors.New("invalid command")
)

// Device structs represents a connected device and can be used to execute
// commands.
type Device struct {
	client *Client
	device message.Device
	done   chan struct{}
}

func (d *Device) String() string {
	return fmt.Sprintf("%s(%d)", d.device.DeviceName, d.device.DeviceIndex)
}

// Name returns the device name.
func (d *Device) Name() string {
	return d.device.DeviceName
}

// IsSupported returns true if the message type is supported.
func (d *Device) IsSupported(msgtype string) bool {
	for _, dm := range d.device.DeviceMessages {
		if dm == msgtype {
			return true
		}
	}
	return false
}

// Supported returns a list of all supported message types for this device.
func (d *Device) Supported() []string {
	return d.device.DeviceMessages
}

// StopDeviceCmd stops a device from whatever actions it may be taking.
func (d *Device) StopDeviceCmd() error {
	if !d.IsSupported(CommandStopDevice) {
		return ErrUnsupported
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		StopDeviceCmd: &message.Device{
			ID:          id,
			DeviceIndex: d.device.DeviceIndex,
		},
	})
}

// RawCmd sends a raw byte string to a device.
func (d *Device) RawCmd(cmd []byte) error {
	if !d.IsSupported(CommandRaw) {
		return ErrUnsupported
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		RawCmd: &message.RawCmd{
			ID:      id,
			Command: cmd,
		},
	})
}

// SingleMotorVibrateCmd causes a toy that supports vibration to run at a
// certain speed. In order to abstract the dynamic range of different toys, the
// value sent is a float with a range of [0.0-1.0].
func (d *Device) SingleMotorVibrateCmd(spd float64) error {
	if !d.IsSupported(CommandSingleMotorVibrate) {
		return ErrUnsupported
	}
	if spd < 0 || spd > 1 {
		return ErrInvalidSpeed
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		SingleMotorVibrateCmd: &message.SingleMotorVibrateCmd{
			ID:    id,
			Speed: spd,
		},
	})
}

// KiirooCmd causes a toy that supports Kiiroo style commands to run whatever
// event may be related.
func (d *Device) KiirooCmd(cmd int) error {
	if !d.IsSupported(CommandKiiroo) {
		return ErrUnsupported
	}
	if cmd < 0 || cmd > 4 {
		return ErrInvalidCmd
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		KiirooCmd: &message.KiirooCmd{
			ID:      id,
			Command: cmd,
		},
	})
}

// FleshlightLaunchFW12Cmd causes a toy that supports Fleshlight Launch
// (Firmware Version 1.2) style commands to run whatever event may be related.
func (d *Device) FleshlightLaunchFW12Cmd(pos, spd int) error {
	if !d.IsSupported(CommandFleshlightLaunchFW12) {
		return ErrUnsupported
	}
	if pos < 0 || pos > 99 {
		return ErrInvalidPosition
	}
	if spd < 0 || spd > 99 {
		return ErrInvalidSpeed
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		FleshlightLaunchFW12Cmd: &message.FleshlightLaunchFW12Cmd{
			ID:       id,
			Position: pos,
			Speed:    spd,
		},
	})
}

// LovenseCmd causes a toy that supports Lovense style commands to run whatever
// event may be related.
func (d *Device) LovenseCmd(cmd string) error {
	if !d.IsSupported(CommandLovense) {
		return ErrUnsupported
	}
	id := d.client.counter.Generate()
	return d.client.sendMessage(id, message.OutgoingMessage{
		LovenseCmd: &message.LovenseCmd{
			ID:      id,
			Command: cmd,
		},
	})
}

// Disconnected returns a receiving channel, that is closed when the device is
// removed.
func (d *Device) Disconnected() <-chan struct{} {
	return d.done
}
