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
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func ExampleDeviceConfig_String() {
	n := DeviceConfig{
		Name:                   "example",
		SeekWindow:             4 * Kibibyte,
		SeekTime:               10 * time.Millisecond,
		ReadBytesPerSecond:     100 * Mebibyte,
		WriteBytesPerSecond:    100 * Mebibyte,
		AllocateBytesPerSecond: 4096 * 100 * Mebibyte,
		RequestReorderMaxDelay: 100 * time.Microsecond,
		FsyncStrategy:          WriteBackCachedFsync,
		WriteStrategy:          FastWrite,
		MetadataOpTime:         10 * time.Millisecond,
	}

	fmt.Println(n.String())
	// Output:
	// example:
	//   SeekWindow             4.10KB (4096)
	//   SeekTime               10ms
	//   ReadBytesPerSecond     104.86MB (104857600)
	//   WriteBytesPerSecond    104.86MB (104857600)
	//   AllocateBytesPerSecond 429.50GB (429496729600)
	//   RequestReorderMaxDelay 100µs
	//   FsyncStrategy          WriteBackCachedFsync
	//   WriteStrategy          FastWrite
	//   MetadataOpTime         10ms

}

func TestComputeTimeFromThroughput(t *testing.T) {
	cases := []struct {
		numBytes       NumBytes
		bytesPerSecond NumBytes
		duration       time.Duration
	}{
		{1, 1, 1 * time.Second},
		{0, 1, 0 * time.Second},
		{1, 1000, 1 * time.Millisecond},
		{1000, 1, 1000 * time.Second},
		{3, 9, 333333333 * time.Nanosecond},
	}

	for _, c := range cases {
		if got, want := computeTimeFromThroughput(c.numBytes, c.bytesPerSecond), c.duration; got != want {
			t.Errorf("computeTimeFromThroughput(%d, %d) = %s, want %s",
				c.numBytes, c.bytesPerSecond, got, want)
		}
	}
}

func TestComputeBytesFromTime(t *testing.T) {
	cases := []struct {
		duration       time.Duration
		bytesPerSecond NumBytes
		want           NumBytes
	}{
		{time.Second, 1, 1},
		{time.Second, 1000, 1000},
		{-time.Second, 100, 0},
		{-time.Second, 0, 0},
		{1500 * time.Millisecond, 1000, 1500},
	}

	for _, c := range cases {
		if got, want := computeBytesFromTime(c.duration, c.bytesPerSecond), c.want; got != want {
			t.Errorf("computeBytesFromTime(%s, %d) = %d, want %d", c.duration, c.bytesPerSecond, got, want)
		}
	}
}

func TestFsyncStrategy_String(t *testing.T) {
	cases := []struct {
		fsyncStrategy FsyncStrategy
		want          string
	}{
		{NoFsync, "NoFsync"},
		{DumbFsync, "DumbFsync"},
		{WriteBackCachedFsync, "WriteBackCachedFsync"},
		{12345, "unknown fsync strategy"},
	}

	for _, c := range cases {
		if got, want := c.fsyncStrategy.String(), c.want; got != want {
			t.Errorf("%d.String() = %s, want %s", c.fsyncStrategy, got, want)
		}
	}
}

func TestParseFsyncStrategyFromString(t *testing.T) {
	cases := []struct {
		strFsyncStrategy string
		want             FsyncStrategy
		shouldErr        bool
	}{
		{"nOFsyNc", NoFsync, false},
		{"no", NoFsync, false},
		{"none", NoFsync, false},
		{"DUmbFsyNc", DumbFsync, false},
		{"dumb", DumbFsync, false},
		{"WriTeBaCkCacHedFsync", WriteBackCachedFsync, false},
		{"wbc", WriteBackCachedFsync, false},
		{"asdfasdf", 0, true},
	}

	for _, c := range cases {
		got, err := ParseFsyncStrategyFromString(c.strFsyncStrategy)
		var expectedErr error
		if c.shouldErr {
			expectedErr = errors.New("expected an error")
		}

		if got != c.want {
			t.Errorf("ParseFsyncStrategyFromString(%s) = %s, want %s", c.strFsyncStrategy, got, c.want)
		}

		if c.shouldErr != (err != nil) {
			t.Errorf("ParseFsyncStrategyFromString(%s) = _, %v, want _, %v", c.strFsyncStrategy, err, expectedErr)
		}
	}
}

func TestWriteStrategy_String(t *testing.T) {
	cases := []struct {
		writeStrategy WriteStrategy
		want          string
	}{
		{FastWrite, "FastWrite"},
		{SimulateWrite, "SimulateWrite"},
		{12345, "unknown write strategy"},
	}

	for _, c := range cases {
		if got, want := c.writeStrategy.String(), c.want; got != want {
			t.Errorf("%d.String() = %s, want %s", c.writeStrategy, got, want)
		}
	}
}

func TestParseWriteStrategyFromString(t *testing.T) {
	cases := []struct {
		strWriteStrategy string
		want             WriteStrategy
		shouldErr        bool
	}{
		{"fAstWrite", FastWrite, false},
		{"fast", FastWrite, false},
		{"sImUlAte", SimulateWrite, false},
		{"sImUlateWrite", SimulateWrite, false},
		{"asdfasdf", 0, true},
	}

	for _, c := range cases {
		got, err := ParseWriteStrategyFromString(c.strWriteStrategy)
		var expectedErr error
		if c.shouldErr {
			expectedErr = errors.New("expected an error")
		}

		if got != c.want {
			t.Errorf("ParseWriteStrategyFromString(%s) = %s, want %s", c.strWriteStrategy, got, c.want)
		}

		if c.shouldErr != (err != nil) {
			t.Errorf("ParseWriteStrategyFromString(%s) = _, %v, want _, %v", c.strWriteStrategy, err, expectedErr)
		}
	}
}

func TestParseDeviceConfigsFromJSON(t *testing.T) {
	cases := []struct {
		jsonDeviceConfig string
		want             []*DeviceConfig
		shouldErr        bool
	}{
		{"", nil, true},
		{"asdfasdf", nil, true},
		{"{}", nil, true},
		{`{"unrecognisedfield":"test"}`, nil, true},
		{`[{"Name":"invalidvalue"}]`, nil, true},
		{
			`{
			  "Name": "7200",
			  "SeekWindow": "4KiB",
			  "SeekTime": "10ms",
			  "ReadBytesPerSecond": "100MiB",
			  "WriteBytesPerSecond": "123KiB",
			  "AllocateBytesPerSecond": "100B",
			  "RequestReorderMaxDelay": "100us",
			  "FsyncStrategy": "wbc",
			  "WriteStrategy": "fastwrite",
			  "MetadataOpTime": "123s"
			}`,
			nil,
			true,
		},
		{
			`[{
			  "Name": "7200",
			  "SeekWindow": "4KiB",
			  "SeekTime": "10ms",
			  "ReadBytesPerSecond": "100aiB",
			  "WriteBytesPerSecond": "123KiB",
			  "AllocateBytesPerSecond": "100B",
			  "RequestReorderMaxDelay": "100us",
			  "FsyncStrategy": "wbc",
			  "WriteStrategy": "fastwrite",
			  "MetadataOpTime": "123s"
			}]`,
			nil,
			true,
		},
		{
			`[{
			  "Name": "7200",
			  "SeekWindow": "4KiB",
			  "SeekTime": "10ms",
			  "ReadBytesPerSecond": "100MiB",
			  "WriteBytesPerSecond": "123KiB",
			  "AllocateBytesPerSecond": "100B",
			  "RequestReorderMaxDelay": "100us",
			  "FsyncStrategy": "wbc",
			  "WriteStrategy": "fastwrite",
			  "MetadataOpTime": "123s"
			}]`,
			[]*DeviceConfig{{
				Name:                   "7200",
				SeekWindow:             4 * Kibibyte,
				SeekTime:               10 * time.Millisecond,
				ReadBytesPerSecond:     100 * Mebibyte,
				WriteBytesPerSecond:    123 * Kibibyte,
				AllocateBytesPerSecond: 100 * Byte,
				RequestReorderMaxDelay: 100 * time.Microsecond,
				FsyncStrategy:          WriteBackCachedFsync,
				WriteStrategy:          FastWrite,
				MetadataOpTime:         123 * time.Second,
			}},
			false,
		},
		{
			`[{
			  "Name": "7200",
			  "SeekWindow": "4KiB",
			  "SeekTime": "10ms",
			  "ReadBytesPerSecond": "100MiB",
			  "WriteBytesPerSecond": "123KiB",
			  "AllocateBytesPerSecond": "100B",
			  "RequestReorderMaxDelay": "100us",
			  "FsyncStrategy": "wbc",
			  "WriteStrategy": "fastwrite",
			  "MetadataOpTime": "123s"
			},
			{
			  "Name": "chicken",
			  "SeekWindow": "0B",
			  "SeekTime": "1ms",
			  "ReadBytesPerSecond": "1B",
			  "WriteBytesPerSecond": "1B",
			  "AllocateBytesPerSecond": "1B",
			  "RequestReorderMaxDelay": "1us",
			  "FsyncStrategy": "no",
			  "WriteStrategy": "simulate",
			  "MetadataOpTime": "0s"
			}]`,
			[]*DeviceConfig{
				{
					Name:                   "7200",
					SeekWindow:             4 * Kibibyte,
					SeekTime:               10 * time.Millisecond,
					ReadBytesPerSecond:     100 * Mebibyte,
					WriteBytesPerSecond:    123 * Kibibyte,
					AllocateBytesPerSecond: 100 * Byte,
					RequestReorderMaxDelay: 100 * time.Microsecond,
					FsyncStrategy:          WriteBackCachedFsync,
					WriteStrategy:          FastWrite,
					MetadataOpTime:         123 * time.Second,
				},
				{
					Name:                   "chicken",
					SeekWindow:             0,
					SeekTime:               1 * time.Millisecond,
					ReadBytesPerSecond:     1 * Byte,
					WriteBytesPerSecond:    1 * Byte,
					AllocateBytesPerSecond: 1 * Byte,
					RequestReorderMaxDelay: 1 * time.Microsecond,
					FsyncStrategy:          NoFsync,
					WriteStrategy:          SimulateWrite,
					MetadataOpTime:         0,
				},
			},
			false,
		},
	}

	for _, c := range cases {
		got, err := ParseDeviceConfigsFromJSON([]byte(c.jsonDeviceConfig))

		if c.shouldErr && err == nil {
			t.Errorf("ParseDeviceConfigsFromJSON(%s) = %s, should error", c.jsonDeviceConfig, got)
		} else if !c.shouldErr {
			if err != nil {
				t.Errorf("ParseDeviceConfigsFromJSON(%s) error: %s, want %s", c.jsonDeviceConfig, err, c.want)
			} else if want := c.want; !reflect.DeepEqual(got, want) {
				t.Errorf("ParseDeviceConfigsFromJSON(%s) = %s, want %s", c.jsonDeviceConfig, got, want)
			}
		}
	}
}
