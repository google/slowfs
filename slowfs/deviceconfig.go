package slowfs

import (
	"time"
)

// DeviceConfig is used to describe how a physical medium acts (e.g. rotational hard drive).
type DeviceConfig struct {
	// SeekWindow describes how many bytes ahead in a file we can access before considering
	// it a seek.
	SeekWindow int64

	// SeekTime denotes the average time of a seek.
	SeekTime time.Duration

	// ReadBytesPerSecond denotes how many bytes we can read per second.
	ReadBytesPerSecond int64

	// ReadBytesPerSecond denotes how many bytes we can write per second.
	WriteBytesPerSecond int64
}

// HardDriveDeviceConfig is a basic model of a 7200rpm hard disk.
var HardDriveDeviceConfig = DeviceConfig{
	SeekWindow:          4 * Kibibyte,
	SeekTime:            10 * time.Millisecond,
	ReadBytesPerSecond:  100 * Mebibyte,
	WriteBytesPerSecond: 100 * Mebibyte,
}
