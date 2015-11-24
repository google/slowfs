package scheduler

import (
	"slowfs/slowfs"
	"time"
)

var startTime time.Time

var basicDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
}

var writeBackCacheDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.WriteBackCachedFsync,
}

var readWriteAsymmetricDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     10 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
}

var notNiceNumbersDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             98 * slowfs.Byte,
	SeekTime:               3*time.Millisecond + 44*time.Microsecond,
	ReadBytesPerSecond:     83 * slowfs.Byte,
	WriteBytesPerSecond:    37 * slowfs.Byte,
	RequestReorderMaxDelay: 6*time.Millisecond + 244*time.Microsecond,
	FsyncStrategy:          slowfs.NoFsync,
}
