/*
Package message contains types and handlers for Buttplug messages.
*/
package message

const (
	// LogLevelOff ...
	LogLevelOff = "Off"
	// LogLevelFatal ...
	LogLevelFatal = "Fatal"
	// LogLevelError ...
	LogLevelError = "Error"
	// LogLevelWarn ...
	LogLevelWarn = "Warn"
	// LogLevelInfo ...
	LogLevelInfo = "Info"
	// LogLevelDebug ...
	LogLevelDebug = "Debug"
	// LogLevelTrace ...
	LogLevelTrace = "Trace"
)

// IncomingMessages list of messages send from a Buttplug server.
type IncomingMessages []IncomingMessage

// OutgoingMessages list of messages send to a Buttplug server.
type OutgoingMessages []OutgoingMessage

// IncomingMessage contains all messages a Buttplug server can send.
type IncomingMessage struct {
	Ok    *Empty `json:"Ok,omitempty"`
	Error *Error `json:"Error,omitempty"`
	Test  *Test  `json:"Test,omitempty"`
	Log   *Log   `json:"Log,omitempty"`

	ServerInfo       *ServerInfo `json:"ServerInfo,omitempty"`
	ScanningFinished *Empty      `json:"ScanningFinished,omitempty"`

	DeviceList    *DeviceList `json:"DeviceList,omitempty"`
	DeviceAdded   *Device     `json:"DeviceAdded,omitempty"`
	DeviceRemoved *Device     `json:"DeviceRemoved,omitempty"`
}

// Message returns the id and message.
func (m IncomingMessage) Message() (id uint32, v interface{}) {
	switch true {
	case m.Ok != nil:
		return m.Ok.ID, *m.Ok
	case m.Error != nil:
		return m.Error.ID, *m.Error
	case m.Test != nil:
		return m.Test.ID, *m.Test
	case m.Log != nil:
		return m.Log.ID, *m.Log
	case m.ServerInfo != nil:
		return m.ServerInfo.ID, *m.ServerInfo
	case m.ScanningFinished != nil:
		return m.ScanningFinished.ID, *m.ScanningFinished
	case m.DeviceList != nil:
		return m.DeviceList.ID, *m.DeviceList
	case m.DeviceAdded != nil:
		return m.DeviceAdded.ID, *m.DeviceAdded
	case m.DeviceRemoved != nil:
		return m.DeviceRemoved.ID, *m.DeviceRemoved
	}
	return 0, nil
}

// OutgoingMessage contains all messages a Buttplug server can receive.
type OutgoingMessage struct {
	Ping       *Empty      `json:"Ping,omitempty"`
	Test       *Test       `json:"Test,omitempty"`
	RequestLog *RequestLog `json:"RequestLog,omitempty"`

	RequestServerInfo *RequestServerInfo `json:"RequestServerInfo,omitempty"`

	StartScanning     *Empty `json:"StartScanning,omitempty"`
	StopScanning      *Empty `json:"StopScanning,omitempty"`
	RequestDeviceList *Empty `json:"RequestDeviceList,omitempty"`

	StopDeviceCmd  *Device `json:"StopDeviceCmd,omitempty"`
	StopAllDevices *Empty  `json:"StopAllDevices,omitempty"`

	RawCmd                  *RawCmd                  `json:"RawCmd,omitempty"`
	SingleMotorVibrateCmd   *SingleMotorVibrateCmd   `json:"SingleMotorVibrateCmd,omitempty"`
	KiirooCmd               *KiirooCmd               `json:"KiirooCmd,omitempty"`
	FleshlightLaunchFW12Cmd *FleshlightLaunchFW12Cmd `json:"FleshlightLaunchFW12Cmd,omitempty"`
	LovenseCmd              *LovenseCmd              `json:"LovenseCmd,omitempty"`
	VorzeA10CycloneCmd      *VorzeA10CycloneCmd      `json:"VorzeA10CycloneCmd,omitempty"`
}

// Empty message is used for all request and responses without additional
// properties.
type Empty struct {
	// The ID the message.
	ID uint32 `json:"Id"`
}

// Error signifies that the previous message sent by the client caused
// some sort of parsing or processing error on the server.
type Error struct {
	// The ID of the client message that this reply is in response to.
	ID uint32 `json:"Id"`
	// Message describing the error that happened on the server.
	ErrorMessage string
}

// Test message is used for development and testing purposes. Sending a Test
// message with a string to the server will cause the server to return a Test
// message. If the string is "Error", the server will return an error message
// instead.
type Test struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// String to echo back from server.
	TestString string
}

// RequestLog requests that the server send internal log messages.
type RequestLog struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// The highest level of message to receive.
	LogLevel string
}

// Log message from the server.
type Log struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// The level of the log message.
	LogLevel string
	// Message
	LogMessage string
}

// RequestServerInfo register with the server, and request info from the
// server.
type RequestServerInfo struct {
	// The ID of the client message that this reply is in response to.
	ID uint32 `json:"Id"`
	// Name of the client, for the server to use for UI if needed.
	ClientName string
}

// ServerInfo contains information about the server name (optional), template
// version, and ping time expectations.
type ServerInfo struct {
	// The ID of the client message that this reply is in response to.
	ID uint32 `json:"Id"`
	// Name of the server.
	ServerName string
	// Message template version of the server software.
	MessageVersion uint32
	// Major version of the server software.
	MajorVersion uint32
	// Minor version of the server software.
	MinorVersion uint32
	// Build version of the server software.
	BuildVersion uint32
	// Maximum internal for pings from the client, in milliseconds.
	MaxPingTime uint32
}

// DeviceList is a server reply to a client request for a device list.
type DeviceList struct {
	// The ID of the client message that this reply is in response to.
	ID uint32 `json:"Id"`
	// Array of device objects
	Devices []Device
}

// Device ...
type Device struct {
	// Message ID (not used when received in a device list.)
	ID uint32 `json:"Id,omitempty"`
	// Descriptive name of the device.
	DeviceName string
	// Index used to identify the device when sending Device Messages.
	DeviceIndex uint32
	// Type names of Device Messages that the device will accept.
	DeviceMessages []string `json:"DeviceMessages,omitempty"`
}

// RawCmd used to send a raw byte string to a device.
type RawCmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Command to execute.
	Command []byte
}

// SingleMotorVibrateCmd causes a toy that supports vibration to run at a
// certain speed.
type SingleMotorVibrateCmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Vibration speed, with a range of [0.0-1.0]
	Speed float64
}

// KiirooCmd causes a toy that supports Kiiroo style commands to run whatever
// event may be related.
type KiirooCmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Command (range [0-4])
	Command int
}

// FleshlightLaunchFW12Cmd causes a toy that supports Fleshlight Launch
// (Firmware Version 1.2) style commands to run whatever event may be related.
type FleshlightLaunchFW12Cmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Position to move to, range 0-99.
	Position int
	// Speed to move at, range 0-99.
	Speed int
}

// LovenseCmd causes a toy that supports Lovense style commands to run whatever
// event may be related.
type LovenseCmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Command for Lovense toys. Must be a valid Lovense command accessible
	// on most of their toys. Implementations should check this for
	// validity.
	Command string
}

// VorzeA10CycloneCmd causes a toy that supports VorzeA10Cyclone style commands
// to run whatever event may be related.
type VorzeA10CycloneCmd struct {
	// Message ID.
	ID uint32 `json:"Id"`
	// Index used to identify the device.
	DeviceIndex uint32
	// Rotation speed command for the Cyclone.
	Speed int
	// True for clockwise rotation (in relation to device facing user),
	// false for Counter-clockwise
	Clockwise bool
}
