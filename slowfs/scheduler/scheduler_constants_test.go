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
	WriteStrategy:          slowfs.SimulateWrite,
}

var fastWriteDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.FastWrite,
}

var writeBackCacheDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     100 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.WriteBackCachedFsync,
	WriteStrategy:          slowfs.FastWrite,
}

var readWriteAsymmetricDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             4 * slowfs.Byte,
	SeekTime:               10 * time.Millisecond,
	ReadBytesPerSecond:     10 * slowfs.Byte,
	WriteBytesPerSecond:    100 * slowfs.Byte,
	RequestReorderMaxDelay: 10 * time.Millisecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.SimulateWrite,
}

var notNiceNumbersDeviceConfig = &slowfs.DeviceConfig{
	SeekWindow:             98 * slowfs.Byte,
	SeekTime:               3*time.Millisecond + 44*time.Microsecond,
	ReadBytesPerSecond:     83 * slowfs.Byte,
	WriteBytesPerSecond:    37 * slowfs.Byte,
	RequestReorderMaxDelay: 6*time.Millisecond + 244*time.Microsecond,
	FsyncStrategy:          slowfs.NoFsync,
	WriteStrategy:          slowfs.SimulateWrite,
}
