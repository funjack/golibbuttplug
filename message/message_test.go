package message

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type MarshalJSONIncoming struct {
	Name string
	JSON string
	Msgs IncomingMessages
}

type MarshalJSONOutgoing struct {
	Name string
	JSON string
	Msgs OutgoingMessages
}

var IncomingJSONCases = []MarshalJSONIncoming{
	{
		Name: "Ok",
		JSON: `[
  {
    "Ok": {
      "Id": 1
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				Ok: &Empty{ID: 1},
			},
		},
	},
	{
		Name: "Error",
		JSON: `[
  {
    "Error": {
      "Id": 0,
      "ErrorMessage": "Server received invalid JSON."
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				Error: &Error{
					ID:           0,
					ErrorMessage: "Server received invalid JSON.",
				},
			},
		},
	},
	{
		Name: "Test",
		JSON: `[
  {
    "Test": {
      "Id": 5,
      "TestString": "Moo"
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				Test: &Test{
					ID:         5,
					TestString: "Moo",
				},
			},
		},
	},
	{
		Name: "Log",
		JSON: `[
  {
    "Log": {
      "Id": 0,
      "LogLevel": "Trace",
      "LogMessage": "This is a Log Message."
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				Log: &Log{
					ID:         0,
					LogLevel:   LogLevelTrace,
					LogMessage: "This is a Log Message.",
				},
			},
		},
	},
	{
		Name: "ServerInfo",
		JSON: `[
  {
    "ServerInfo": {
      "Id": 1,
      "ServerName": "Test Server",
      "MessageVersion": 1,
      "MajorVersion": 1,
      "MinorVersion": 0,
      "BuildVersion": 0,
      "MaxPingTime": 100
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				ServerInfo: &ServerInfo{
					ID:             1,
					ServerName:     "Test Server",
					MessageVersion: 1,
					MajorVersion:   1,
					MinorVersion:   0,
					BuildVersion:   0,
					MaxPingTime:    100,
				},
			},
		},
	},
	{
		Name: "ScanningFinished",
		JSON: `[
  {
    "ScanningFinished": {
      "Id": 0
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				ScanningFinished: &Empty{
					ID: 0,
				},
			},
		},
	},
	{
		Name: "DeviceList",
		JSON: `[
  {
    "DeviceList": {
      "Id": 1,
      "Devices": [
        {
          "DeviceName": "TestDevice 1",
          "DeviceIndex": 0,
          "DeviceMessages": ["SingleMotorVibrateCmd", "RawCmd", "KiirooCmd", "StopDeviceCmd"]
        },
        {
          "DeviceName": "TestDevice 2",
          "DeviceIndex": 1,
          "DeviceMessages": ["SingleMotorVibrateCmd", "LovenseCmd", "StopDeviceCmd"]
        }
      ]
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				DeviceList: &DeviceList{
					ID: 1,
					Devices: []Device{
						{
							DeviceName:  "TestDevice 1",
							DeviceIndex: 0,
							DeviceMessages: []string{
								"SingleMotorVibrateCmd",
								"RawCmd",
								"KiirooCmd",
								"StopDeviceCmd",
							},
						},
						{
							DeviceName:  "TestDevice 2",
							DeviceIndex: 1,
							DeviceMessages: []string{
								"SingleMotorVibrateCmd",
								"LovenseCmd",
								"StopDeviceCmd",
							},
						},
					},
				},
			},
		},
	},
	{
		Name: "DeviceAdded",
		JSON: `[
  {
    "DeviceAdded": {
      "Id": 0,
      "DeviceName": "TestDevice 1",
      "DeviceIndex": 0,
      "DeviceMessages": ["SingleMotorVibrateCmd", "RawCmd", "KiirooCmd", "StopDeviceCmd"]
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				DeviceAdded: &Device{
					ID:          0,
					DeviceName:  "TestDevice 1",
					DeviceIndex: 0,
					DeviceMessages: []string{
						"SingleMotorVibrateCmd",
						"RawCmd",
						"KiirooCmd",
						"StopDeviceCmd",
					},
				},
			},
		},
	},
	{
		Name: "DeviceRemoved",
		JSON: `[
  {
    "DeviceRemoved": {
      "Id": 0,
      "DeviceIndex": 0
    }
  }
]`,
		Msgs: IncomingMessages{
			{
				DeviceRemoved: &Device{
					ID:          0,
					DeviceIndex: 0,
				},
			},
		},
	},
}

var OutgoingJSONCases = []MarshalJSONOutgoing{
	{
		Name: "Ping",
		JSON: `[
  {
    "Ping": {
      "Id": 5
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				Ping: &Empty{
					ID: 5,
				},
			},
		},
	},
	{
		Name: "Test",
		JSON: `[
  {
    "Test": {
      "Id": 5,
      "TestString": "Moo"
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				Test: &Test{
					ID:         5,
					TestString: "Moo",
				},
			},
		},
	},
	{
		Name: "RequestLog",
		JSON: `[
  {
    "RequestLog": {
      "Id": 1,
      "LogLevel": "Warn"
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				RequestLog: &RequestLog{
					ID:       1,
					LogLevel: LogLevelWarn,
				},
			},
		},
	},
	{
		Name: "RequestServerInfo",
		JSON: `[
  {
    "RequestServerInfo": {
      "Id": 1,
      "ClientName": "Test Client"
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				RequestServerInfo: &RequestServerInfo{
					ID:         1,
					ClientName: "Test Client",
				},
			},
		},
	},
	{
		Name: "StartScanning",
		JSON: `[
  {
    "StartScanning": {
      "Id": 1
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				StartScanning: &Empty{
					ID: 1,
				},
			},
		},
	},
	{
		Name: "StopScanning",
		JSON: `[
  {
    "StopScanning": {
      "Id": 1
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				StopScanning: &Empty{
					ID: 1,
				},
			},
		},
	},
	{
		Name: "RequestDeviceList",
		JSON: `[
  {
    "RequestDeviceList": {
      "Id": 1
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				RequestDeviceList: &Empty{
					ID: 1,
				},
			},
		},
	},
	{
		Name: "StopDeviceCmd",
		JSON: `[
  {
    "StopDeviceCmd": {
      "Id": 1,
      "DeviceIndex": 0
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				StopDeviceCmd: &Device{
					ID:          1,
					DeviceIndex: 0,
				},
			},
		},
	},
	{
		Name: "StopAllDevices",
		JSON: `[
  {
    "StopAllDevices": {
      "Id": 1
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				StopAllDevices: &Empty{
					ID: 1,
				},
			},
		},
	},
	{
		Name: "RawCmd",
		JSON: `[
  {
    "RawCmd": {
      "Id": 1,
      "DeviceIndex": 0,
      "Command": [0, 2, 4]
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				RawCmd: &RawCmd{
					ID:          1,
					DeviceIndex: 0,
					Command:     []byte{0x00, 0x02, 0x04},
				},
			},
		},
	},
	{
		Name: "SingleMotorVibrateCmd",
		JSON: `[
  {
    "SingleMotorVibrateCmd": {
      "Id": 1,
      "DeviceIndex": 0,
      "Speed": 0.5
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				SingleMotorVibrateCmd: &SingleMotorVibrateCmd{
					ID:          1,
					DeviceIndex: 0,
					Speed:       0.5,
				},
			},
		},
	},
	{
		Name: "KiirooCmd",
		JSON: `[
  {
    "KiirooCmd": {
      "Id": 1,
      "DeviceIndex": 0,
      "Command": 4
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				KiirooCmd: &KiirooCmd{
					ID:          1,
					DeviceIndex: 0,
					Command:     4,
				},
			},
		},
	},
	{
		Name: "FleshlightLaunchFW12Cmd",
		JSON: `[
  {
    "FleshlightLaunchFW12Cmd": {
      "Id": 1,
      "DeviceIndex": 0,
      "Position": 95,
      "Speed": 90
    }
  }
]`,
		Msgs: OutgoingMessages{
			{
				FleshlightLaunchFW12Cmd: &FleshlightLaunchFW12Cmd{
					ID:          1,
					DeviceIndex: 0,
					Position:    95,
					Speed:       90,
				},
			},
		},
	},
}

func TestMarshallingJSONIncoming(t *testing.T) {
	for _, c := range IncomingJSONCases {
		var imsg IncomingMessages
		if err := marshalJSONInOut(c.JSON, &imsg, c.Msgs); err != nil {
			t.Errorf("case %s: error %v", c.Name, err)
		}
	}
}

func TestMarshallingJSONOutgoing(t *testing.T) {
	for _, c := range OutgoingJSONCases {
		var omsg OutgoingMessages
		if err := marshalJSONInOut(c.JSON, &omsg, c.Msgs); err != nil {
			t.Errorf("case %s: error %v", c.Name, err)
		}
	}
}

func marshalJSONInOut(in string, out, want interface{}) error {
	if err := json.Unmarshal([]byte(in), out); err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}
	ok, err := equalMessages(out, want)
	if err != nil {
		return fmt.Errorf("equals error: %v", err)
	}
	if !ok {
		return fmt.Errorf("unmarshal result does not match")
	}
	marshaled, err := json.Marshal(&want)
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}
	if err := json.Unmarshal(marshaled, out); err != nil {
		return fmt.Errorf("error unmarshalling marshaled result: %v", err)
	}
	ok, err = equalMessages(out, want)
	if err != nil {
		return fmt.Errorf("equals error: %v", err)
	}
	if !ok {
		return fmt.Errorf("unmarshal result does not match")
	}
	return nil
}

func equalMessages(a, b interface{}) (bool, error) {
	switch x := a.(type) {
	default:
		return false, fmt.Errorf("unexpected type %T", x)
	case *IncomingMessages:
		if y, ok := b.(IncomingMessages); ok {
			return (*x).Equals(y), nil
		}
		return false, fmt.Errorf("unexpected type")
	case *OutgoingMessages:
		if y, ok := b.(OutgoingMessages); ok {
			return (*x).Equals(y), nil
		}
		return false, fmt.Errorf("unexpected type")
	}
}

// Equals returns true if the pointer reference values match.
func (p IncomingMessages) Equals(v IncomingMessages) bool {
	if len(p) != len(v) {
		return false
	}
	for i := range p {
		if !p[i].Equals(&v[i]) {
			return false
		}
	}
	return true
}

// Equals returns true if the pointer reference values match.
func (p OutgoingMessages) Equals(v OutgoingMessages) bool {
	if len(p) != len(v) {
		return false
	}
	for i := range p {
		if !p[i].Equals(&v[i]) {
			return false
		}
	}
	return true
}

// Equals returns true if the pointer references match.
func (p *IncomingMessage) Equals(v *IncomingMessage) bool {
	if p == nil && v == nil {
		return true
	} else if p == nil || v == nil {
		return false
	}

	switch true {
	case p.Ok == nil && v.Ok != nil:
		return false
	case p.Ok != nil && *p.Ok != *v.Ok:
		return false
	case p.Error == nil && v.Error != nil:
		return false
	case p.Error != nil && *p.Error != *v.Error:
		return false
	case p.Test == nil && v.Test != nil:
		return false
	case p.Test != nil && *p.Test != *v.Test:
		return false
	case p.Log == nil && v.Log != nil:
		return false
	case p.Log != nil && *p.Log != *v.Log:
		return false
	case p.ServerInfo == nil && v.ServerInfo != nil:
		return false
	case p.ServerInfo != nil && *p.ServerInfo != *v.ServerInfo:
		return false
	case p.ScanningFinished == nil && v.ScanningFinished != nil:
		return false
	case p.ScanningFinished != nil && *p.ScanningFinished != *v.ScanningFinished:
		return false
	case p.DeviceList == nil && v.DeviceList != nil:
		return false
	case p.DeviceList != nil && !reflect.DeepEqual(*p.DeviceList, *v.DeviceList):
		return false
	case p.DeviceAdded == nil && v.DeviceAdded != nil:
		return false
	case p.DeviceAdded != nil && !reflect.DeepEqual(*p.DeviceAdded, *v.DeviceAdded):
		return false
	case p.DeviceRemoved == nil && v.DeviceRemoved != nil:
		return false
	case p.DeviceRemoved != nil && !reflect.DeepEqual(*p.DeviceRemoved, *v.DeviceRemoved):
		return false
	}
	return true
}

// Equals returns true if the pointer references match.
func (p *OutgoingMessage) Equals(v *OutgoingMessage) bool {
	if p == nil && v == nil {
		return true
	} else if p == nil || v == nil {
		return false
	}
	switch true {
	case p.Ping == nil && v.Ping != nil:
		return false
	case p.Ping != nil && *p.Ping != *v.Ping:
		return false
	case p.Test == nil && v.Test != nil:
		return false
	case p.Test != nil && *p.Test != *v.Test:
		return false
	case p.RequestLog == nil && v.RequestLog != nil:
		return false
	case p.RequestLog != nil && *p.RequestLog != *v.RequestLog:
		return false
	case p.RequestServerInfo == nil && v.RequestServerInfo != nil:
		return false
	case p.RequestServerInfo != nil && *p.RequestServerInfo != *v.RequestServerInfo:
		return false
	case p.StartScanning == nil && v.StartScanning != nil:
		return false
	case p.StartScanning != nil && *p.StartScanning != *v.StartScanning:
		return false
	case p.StopScanning == nil && v.StopScanning != nil:
		return false
	case p.StopScanning != nil && *p.StopScanning != *v.StopScanning:
		return false
	case p.RequestDeviceList == nil && v.RequestDeviceList != nil:
		return false
	case p.RequestDeviceList != nil && *p.RequestDeviceList != *v.RequestDeviceList:
		return false
	case p.StopDeviceCmd == nil && v.StopDeviceCmd != nil:
		return false
	case p.StopDeviceCmd != nil && !reflect.DeepEqual(*p.StopDeviceCmd, *v.StopDeviceCmd):
		return false
	case p.StopAllDevices == nil && v.StopAllDevices != nil:
		return false
	case p.StopAllDevices != nil && *p.StopAllDevices != *v.StopAllDevices:
		return false
	case p.RawCmd == nil && v.RawCmd != nil:
		return false
	case p.RawCmd != nil && !reflect.DeepEqual(*p.RawCmd, *v.RawCmd):
		return false
	case p.SingleMotorVibrateCmd == nil && v.SingleMotorVibrateCmd != nil:
		return false
	case p.SingleMotorVibrateCmd != nil && *p.SingleMotorVibrateCmd != *v.SingleMotorVibrateCmd:
		return false
	case p.KiirooCmd == nil && v.KiirooCmd != nil:
		return false
	case p.KiirooCmd != nil && *p.KiirooCmd != *v.KiirooCmd:
		return false
	case p.FleshlightLaunchFW12Cmd == nil && v.FleshlightLaunchFW12Cmd != nil:
		return false
	case p.FleshlightLaunchFW12Cmd != nil && *p.FleshlightLaunchFW12Cmd != *v.FleshlightLaunchFW12Cmd:
		return false
	case p.LovenseCmd == nil && v.LovenseCmd != nil:
		return false
	case p.LovenseCmd != nil && *p.LovenseCmd != *v.LovenseCmd:
		return false
	}
	return true
}
