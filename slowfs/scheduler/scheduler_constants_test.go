// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"slowfs/slowfs"
	"slowfs/slowfs/units"
	"time"
)

var startTime time.Time

var basicDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * units.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * units.Byte,
	WriteBytesPerSecond:    100 * units.Byte,
	AllocateBytesPerSecond: 1000 * units.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.SimulateWrite,
	MetadataOpTime:         80 * time.Millisecond,
}

var fastWriteDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * units.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * units.Byte,
	WriteBytesPerSecond:    100 * units.Byte,
	AllocateBytesPerSecond: 1000 * units.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.FastWrite,
	MetadataOpTime:         80 * time.Millisecond,
}

var writeBackCacheDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * units.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * units.Byte,
	WriteBytesPerSecond:    100 * units.Byte,
	AllocateBytesPerSecond: 1000 * units.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.WriteBackCachedFsync,
	WriteStrategy:          slowfs.FastWrite,
	MetadataOpTime:         80 * time.Millisecond,
}

var readWriteAsymmetricDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * units.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     10 * units.Byte,
	WriteBytesPerSecond:    100 * units.Byte,
	AllocateBytesPerSecond: 1000 * units.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.SimulateWrite,
	MetadataOpTime:         80 * time.Millisecond,
}

var notNiceNumbersDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             98 * units.Byte,
	SeekTime:               3*time.Millisecond + 44*time.Microsecond,
	ReadBytesPerSecond:     83 * units.Byte,
	WriteBytesPerSecond:    37 * units.Byte,
	AllocateBytesPerSecond: 1000 * units.Byte,
	RequestReorderMaxDelay: 6*time.Millisecond + 244*time.Microsecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.SimulateWrite,
	MetadataOpTime:         80 * time.Millisecond,
}
