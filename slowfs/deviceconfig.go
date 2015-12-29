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

package slowfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slowfs/slowfs/units"
	"strings"
	"time"
)

// FsyncStrategy indicates which strategy to use for fsync simulation.
type FsyncStrategy int

const (
	// NoFsync indicates a strategy where fsync takes zero time.
	NoFsync FsyncStrategy = iota
	// DumbFsync indicates a strategy where fsync takes ten seek times (chosen arbitrarily).
	DumbFsync
	// WriteBackCachedFsync indicates a simulation of write back cache. This means writes will take
	// very little time, and writing back that data to disk will be simulated to happen during spare
	// IO time. When fsync is called on a file, how much unwritten data remaining for that file
	// determines how long the fsync takes.
	WriteBackCachedFsync
)

func (f FsyncStrategy) String() string {
	switch f {
	case NoFsync:
		return "NoFsync"
	case DumbFsync:
		return "DumbFsync"
	case WriteBackCachedFsync:
		return "WriteBackCachedFsync"
	default:
		return "unknown fsync strategy"
	}
}

// ParseFsyncStrategyFromString parses a FsyncStrategy from a string. There can be multiple ways to
// specify each FsyncStrategy (e.g. nofsync, none, and no all mean 'NoFsync'). This function is
// case insensitive.
func ParseFsyncStrategyFromString(s string) (FsyncStrategy, error) {
	switch strings.ToLower(s) {
	case "nofsync", "none", "no":
		return NoFsync, nil
	case "dumbfsync", "dumb":
		return DumbFsync, nil
	case "writebackcachedfsync", "writebackcache", "wbc":
		return WriteBackCachedFsync, nil
	default:
		return 0, fmt.Errorf("unknown fsync strategy %s", s)
	}
}

// WriteStrategy indicates which strategy to use for write simulation.
type WriteStrategy int

const (
	// FastWrite means writes will take zero time, as if they were cached.
	// Useful in conjunction with WriteBackCachedFsync
	FastWrite WriteStrategy = iota
	// SimulateWrite means writes will act in the same way as reads -- that is,
	// seeking if non-sequential, and being written out at the speed specified
	// by WriteBytesPerSecond.
	SimulateWrite
)

func (w WriteStrategy) String() string {
	switch w {
	case FastWrite:
		return "FastWrite"
	case SimulateWrite:
		return "SimulateWrite"
	default:
		return "unknown write strategy"
	}
}

// ParseWriteStrategyFromString parses a WriteStrategy from the given string. This function is
// case insensitive, and also accepts synonyms for each WriteStrategy. For example, fastwrite and
// fast both map to the FastWrite strategy.
func ParseWriteStrategyFromString(s string) (WriteStrategy, error) {
	switch strings.ToLower(s) {
	case "fastwrite", "fast":
		return FastWrite, nil
	case "simulatewrite", "simulate":
		return SimulateWrite, nil
	default:
		return 0, fmt.Errorf("unknown write strategy %s", s)
	}
}

// DeviceConfig is used to describe how a physical medium acts (e.g. rotational hard drive).
type DeviceConfig struct {
	// Name is the name of this configuration. This is used for selecting on the command line which
	// configuration to use.
	Name string

	// SeekWindow describes how many bytes ahead in a file we can access before considering
	// it a seek.
	SeekWindow units.NumBytes

	// SeekTime denotes the average time of a seek.
	SeekTime time.Duration

	// ReadBytesPerSecond denotes how many bytes we can read per second.
	ReadBytesPerSecond units.NumBytes

	// ReadBytesPerSecond denotes how many bytes we can write per second.
	WriteBytesPerSecond units.NumBytes

	// AllocateBytesPerSecond denotes how many bytes we can allocate using
	// fallocate per second.
	AllocateBytesPerSecond units.NumBytes

	// RequestReorderMaxDelay denotes how much later a request can be by timestamp after a previous
	// one and still be reordered before it.
	RequestReorderMaxDelay time.Duration

	// FsyncStrategy denotes which algorithm to use for modeling fsync.
	FsyncStrategy FsyncStrategy

	// WriteStrategy denotes which algorithm to use for modeling writes.
	WriteStrategy WriteStrategy

	// MetadataOpTime denotes how long metadata operations (like chmod, chown, etc) should take.
	MetadataOpTime time.Duration
}

func (dc *DeviceConfig) String() string {
	return fmt.Sprintf(`%s:
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s
  %-22s %s`,
		dc.Name, "SeekWindow", dc.SeekWindow, "SeekTime", dc.SeekTime,
		"ReadBytesPerSecond", dc.ReadBytesPerSecond, "WriteBytesPerSecond", dc.WriteBytesPerSecond,
		"AllocateBytesPerSecond", dc.AllocateBytesPerSecond, "RequestReorderMaxDelay", dc.RequestReorderMaxDelay,
		"FsyncStrategy", dc.FsyncStrategy, "WriteStrategy", dc.WriteStrategy, "MetadataOpTime", dc.MetadataOpTime)
}

func parseDeviceConfig(obj map[string]interface{}) (*DeviceConfig, error) {
	var dc DeviceConfig

	missingFields := map[string]struct{}{
		"Name":                   {},
		"SeekWindow":             {},
		"SeekTime":               {},
		"ReadBytesPerSecond":     {},
		"WriteBytesPerSecond":    {},
		"AllocateBytesPerSecond": {},
		"RequestReorderMaxDelay": {},
		"FsyncStrategy":          {},
		"WriteStrategy":          {},
		"MetadataOpTime":         {},
	}

	for k, v := range obj {
		if _, ok := missingFields[k]; !ok {
			return nil, fmt.Errorf("spurious field %s", k)
		}
		delete(missingFields, k)

		strVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%s: want string type, got %v", k, v)
		}

		var err error
		switch k {
		case "Name":
			dc.Name = strVal
		case "SeekWindow":
			dc.SeekWindow, err = units.ParseNumBytesFromString(strVal)
		case "SeekTime":
			dc.SeekTime, err = time.ParseDuration(strVal)
		case "ReadBytesPerSecond":
			dc.ReadBytesPerSecond, err = units.ParseNumBytesFromString(strVal)
		case "WriteBytesPerSecond":
			dc.WriteBytesPerSecond, err = units.ParseNumBytesFromString(strVal)
		case "AllocateBytesPerSecond":
			dc.AllocateBytesPerSecond, err = units.ParseNumBytesFromString(strVal)
		case "RequestReorderMaxDelay":
			dc.RequestReorderMaxDelay, err = time.ParseDuration(strVal)
		case "FsyncStrategy":
			dc.FsyncStrategy, err = ParseFsyncStrategyFromString(strVal)
		case "WriteStrategy":
			dc.WriteStrategy, err = ParseWriteStrategyFromString(strVal)
		case "MetadataOpTime":
			dc.MetadataOpTime, err = time.ParseDuration(strVal)
		default:
			panic("bug")
		}

		if err != nil {
			return nil, fmt.Errorf("%s: %s", k, err)
		}

	}

	if len(missingFields) != 0 {
		var strFields string
		for k := range missingFields {
			strFields += k + " "
		}
		return nil, fmt.Errorf("missing fields: %s", strFields)
	}

	return &dc, nil
}

// ParseDeviceConfigsFromJSON parses json containing an array of device configs.
func ParseDeviceConfigsFromJSON(data []byte) ([]*DeviceConfig, error) {
	// We can't set required fields or similar, so check for missing fields or spurious fields
	// manually.
	var dcObjs []map[string]interface{}
	err := json.Unmarshal(data, &dcObjs)
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return nil, fmt.Errorf("expected array containing device configs")
	}
	if err != nil {
		return nil, err
	}

	dcs := make([]*DeviceConfig, 0, len(dcObjs))
	for _, dcObj := range dcObjs {
		dc, err := parseDeviceConfig(dcObj)
		if err != nil {
			return nil, fmt.Errorf("error validating device config %v: %s", dcObj, err)
		}
		dcs = append(dcs, dc)
	}

	return dcs, nil
}

// Validate decides whether a device config is valid or not. If a device config has fields that
// don't make sense (like negative delays), it will return an error. If there are field combinations
// that /probably/ don't make sense it will print a warning message.
func (dc *DeviceConfig) Validate() error {
	if dc.SeekWindow < 0 {
		return errors.New("SeekWindow cannot be negative.")
	}
	if dc.SeekTime < 0 {
		return errors.New("SeekTime cannot be negative.")
	}
	if dc.ReadBytesPerSecond <= 0 {
		return errors.New("ReadBytesPerSecond cannot be non-positive.")
	}
	if dc.WriteBytesPerSecond <= 0 {
		return errors.New("WriteBytesPerSecond cannot be non-positive.")
	}
	if dc.AllocateBytesPerSecond <= 0 {
		return errors.New("AllocateBytesPerSecond cannot be non-positive.")
	}
	if dc.RequestReorderMaxDelay < 0 {
		return errors.New("RequestReorderMaxDelay cannot be negative.")
	}
	if dc.RequestReorderMaxDelay > 500*time.Microsecond {
		log.Println("setting RequestReorderMaxDelay to >500us is probably not what you want")
	}
	if dc.MetadataOpTime < 0 {
		return errors.New("MetadataOpTime cannot be negative.")
	}

	if dc.WriteStrategy == SimulateWrite && dc.FsyncStrategy == WriteBackCachedFsync {
		log.Println("setting both simulated writes and write back cache is probably not what you want. " +
			"Write back cache is meant to simulate writes being cached in memory and taking minimal time, " +
			"then being written back to disk later, either during spare IO time or at an fsync.")
	}

	return nil
}

// WriteTime computes how long writing numBytes will take.
func (dc *DeviceConfig) WriteTime(numBytes units.NumBytes) time.Duration {
	return computeTimeFromThroughput(numBytes, dc.WriteBytesPerSecond)
}

// ReadTime computes how long reading numBytes will take.
func (dc *DeviceConfig) ReadTime(numBytes units.NumBytes) time.Duration {
	return computeTimeFromThroughput(numBytes, dc.ReadBytesPerSecond)
}

// AllocateTime computes how long allocating numBytes will take.
func (dc *DeviceConfig) AllocateTime(numBytes units.NumBytes) time.Duration {
	return computeTimeFromThroughput(numBytes, dc.AllocateBytesPerSecond)
}

// WritableBytes computes how many bytes can be written in the given duration.
func (dc *DeviceConfig) WritableBytes(duration time.Duration) units.NumBytes {
	return computeBytesFromTime(duration, dc.WriteBytesPerSecond)
}

// ReadableBytes computes how many bytes can be read in the given duration.
func (dc *DeviceConfig) ReadableBytes(duration time.Duration) units.NumBytes {
	return computeBytesFromTime(duration, dc.ReadBytesPerSecond)
}

func computeTimeFromThroughput(numBytes, bytesPerSecond units.NumBytes) time.Duration {
	return time.Duration(float64(numBytes) / float64(bytesPerSecond) * float64(time.Second))
}

func computeBytesFromTime(duration time.Duration, bytesPerSecond units.NumBytes) units.NumBytes {
	if duration <= 0 {
		return 0
	}
	return units.NumBytes(float64(duration) / float64(time.Second) * float64(bytesPerSecond))
}

// Below follows the list of preset device configurations. If you add configurations, please
// update the tests to Validate() them.

// HDD7200RpmDeviceConfig is a basic model of a 7200rpm hard disk.
var HDD7200RpmDeviceConfig = DeviceConfig{
	Name:                "hdd7200rpm",
	SeekWindow:          4 * units.Kibibyte,
	SeekTime:            10 * time.Millisecond,
	ReadBytesPerSecond:  100 * units.Mebibyte,
	WriteBytesPerSecond: 100 * units.Mebibyte,
	// Default to 4096 times faster than writing, since ext4 block sizes are
	// 4 KiB.
	AllocateBytesPerSecond: 4096 * 100 * units.Mebibyte,
	RequestReorderMaxDelay: 100 * time.Microsecond,
	FsyncStrategy:          WriteBackCachedFsync,
	WriteStrategy:          FastWrite,
	MetadataOpTime:         10 * time.Millisecond,
}
