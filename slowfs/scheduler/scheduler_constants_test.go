package scheduler

import (
	"slowfs/slowfs"
	"time"
)

var startTime time.Time

var basicDeviceConfig = slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
}

var readWriteAsymmetricDeviceConfig = slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     10 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
}
