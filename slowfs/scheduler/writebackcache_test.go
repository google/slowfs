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
	"reflect"
	"slowfs/slowfs"
	"sort"
	"testing"
	"time"
)

func TestWriteBackCache_Write(t *testing.T) {
	cases := []struct {
		path     string
		numBytes int64
		want     int64
	}{{"a", 101, 101}, {"b", 102, 102}, {"c", 0, 0}, {"c", 0, 0}, {"c", 1, 1}, {"c", 5, 6}, {"a", 1, 102}, {"b", 102, 204}}

	writeBackCache := newWriteBackCache(basicDeviceConfig)
	for _, c := range cases {
		writeBackCache.write(c.path, c.numBytes)
		if got, want := writeBackCache.getUnwrittenBytes(c.path), c.want; got != want {
			t.Errorf("getUnwrittenBytes(%s) = %d, want %d", c.path, got, want)
		}
	}
}

func TestWriteBackCache_Close(t *testing.T) {
	cases := []struct {
		path     string
		numBytes int64
		want     int64
	}{{"a", 101, 101}, {"b", 102, 203}, {"c", 0, 203}, {"c", 0, 203}, {"c", 1, 204}, {"c", 5, 209}, {"a", 1, 210}, {"b", 102, 312}}

	writeBackCache := newWriteBackCache(basicDeviceConfig)
	for _, c := range cases {
		writeBackCache.write(c.path, c.numBytes)
		writeBackCache.close(c.path)

		if got, want := writeBackCache.getUnwrittenBytes(c.path), int64(0); got != want {
			t.Errorf("getUnwrittenBytes(%s) = %d, want %d", c.path, got, want)
		}
		if got, want := writeBackCache.orphanedUnwrittenBytes, c.want; got != want {
			t.Errorf("orphanedUnwrittenBytes = %d, want %d", writeBackCache.orphanedUnwrittenBytes, want)
		}
	}
}

func TestWriteBackCache_WriteBack(t *testing.T) {
	type writeInvocation struct {
		path        string
		numBytes    int64
		shouldClose bool
	}
	type writeBackInvocation struct {
		duration      time.Duration
		wantRemaining int64
	}
	cases := []struct {
		desc       string
		writes     []writeInvocation
		writeBacks []writeBackInvocation
	}{
		{
			"nothing",
			[]writeInvocation{},
			[]writeBackInvocation{
				{0, 0},
				{time.Second, 0},
			},
		},
		{
			"big write back",
			[]writeInvocation{
				{"a", 100, false},
				{"b", 100, true},
				{"b", 100, false},
				{"c", 200, false},
				{"d", 200, true},
				{"d", 200, true},
				{"d", 200, false},
				{"a", 100, false},
				{"b", 100, false},
			},
			[]writeBackInvocation{
				{0, 1300},
				{time.Hour, 0},
			},
		},
		{
			"repeated write backs",
			[]writeInvocation{
				{"a", 20, false},
				{"b", 10, true},
				{"c", 20, false},
				{"d", 10, true},
			},
			[]writeBackInvocation{
				{0, 60},
				{10 * time.Millisecond, 60},
				{5 * time.Millisecond, 60},
				{19 * time.Millisecond, 60},
				{15 * time.Millisecond, 60},
				{20 * time.Millisecond, 59},
				{100 * time.Millisecond, 50},
				{530 * time.Millisecond, 0},
				{500 * time.Millisecond, 0},
			},
		},
	}

	for _, c := range cases {
		writeBackCache := newWriteBackCache(basicDeviceConfig)
		for _, write := range c.writes {
			writeBackCache.write(write.path, write.numBytes)
			if write.shouldClose {
				writeBackCache.close(write.path)
			}
		}

		for _, writeBack := range c.writeBacks {
			writeBackCache.writeBack(writeBack.duration)
			remainingBytes := writeBackCache.orphanedUnwrittenBytes
			for _, bytes := range writeBackCache.unwrittenBytes {
				remainingBytes += bytes
			}
			if got, want := remainingBytes, writeBack.wantRemaining; got != want {
				t.Errorf("fail (%s) writeBack(%s) leaves %d bytes remaining, want %d", c.desc, writeBack.duration, got, want)
			}
		}
	}
}

func TestWriteBackCache_WriteBackBytesForFile(t *testing.T) {
	cases := []struct {
		desc          string
		deviceConfig  *slowfs.DeviceConfig
		numBytes      int64
		duration      time.Duration
		wantDuration  time.Duration
		wantRemaining int64
	}{
		{
			desc:          "no time to seek",
			deviceConfig:  basicDeviceConfig,
			numBytes:      1,
			duration:      9 * time.Millisecond,
			wantDuration:  0 * time.Millisecond,
			wantRemaining: 1,
		},
		{
			desc:          "no time to seek",
			deviceConfig:  basicDeviceConfig,
			numBytes:      6,
			duration:      4 * time.Millisecond,
			wantDuration:  0 * time.Millisecond,
			wantRemaining: 6,
		},
		{
			desc:          "time limited write back",
			deviceConfig:  basicDeviceConfig,
			numBytes:      100,
			duration:      510 * time.Millisecond,
			wantDuration:  510 * time.Millisecond,
			wantRemaining: 50,
		},
		{
			desc:          "non-nice number duration",
			deviceConfig:  basicDeviceConfig,
			numBytes:      97,
			duration:      467 * time.Millisecond,
			wantDuration:  460 * time.Millisecond,
			wantRemaining: 52,
		},
		{
			desc:          "no bytes",
			deviceConfig:  basicDeviceConfig,
			numBytes:      0,
			duration:      100 * time.Millisecond,
			wantDuration:  0 * time.Millisecond,
			wantRemaining: 0,
		},
		{
			desc:          "zero duration",
			deviceConfig:  basicDeviceConfig,
			numBytes:      1,
			duration:      0,
			wantDuration:  0 * time.Millisecond,
			wantRemaining: 1,
		},
		{
			desc:          "byte limited write back",
			deviceConfig:  basicDeviceConfig,
			numBytes:      10,
			duration:      510 * time.Millisecond,
			wantDuration:  110 * time.Millisecond,
			wantRemaining: 0,
		},
		{
			desc:          "odd numbers",
			deviceConfig:  notNiceNumbersDeviceConfig,
			numBytes:      13,
			duration:      137*time.Millisecond + 543*time.Microsecond,
			wantDuration:  111152108 * time.Nanosecond,
			wantRemaining: 9,
		},
	}

	for _, c := range cases {
		writeBackCache := newWriteBackCache(c.deviceConfig)
		writeBackCache.write("a", c.numBytes)

		if got, want := writeBackCache.writeBackBytesForFile("a", c.duration), c.wantDuration; got != want {
			t.Errorf("fail (%s) writeBackBytesForFile(\"a\", %s) = %s, want %s", c.desc, c.duration, got, want)
		}
		if got, want := writeBackCache.getUnwrittenBytes("a"), c.wantRemaining; got != want {
			t.Errorf("fail (%s) getUnwrittenBytes(\"a\") = %d, want %d", c.desc, got, want)
		}
	}
}

func TestInt64Min(t *testing.T) {
	cases := []struct {
		a    int64
		b    int64
		want int64
	}{
		{1, 1, 1},
		{100, -12, -12},
		{100, 101, 100},
		{0, 1, 0},
	}

	for _, c := range cases {
		if got, want := int64Min(c.a, c.b), c.want; got != want {
			t.Errorf("int64Min(%d, %d) = %d, want %d", c.a, c.b, got, want)
		}
	}
}

func TestComputeWritableBytes(t *testing.T) {
	cases := []struct {
		duration       time.Duration
		bytesPerSecond int64
		seekTime       time.Duration
		want           int64
	}{
		{time.Second, 1, 0, 1},
		{time.Second, 1000, 0, 1000},
		{-time.Second, 100, 0, 0},
		{-time.Second, 0, 0, 0},
		{time.Second, 1, time.Second, 0},
		{time.Second, 1000, 500 * time.Millisecond, 500},
		{2 * time.Second, 1000, 500 * time.Millisecond, 1500},
		{-time.Second, 100, 500 * time.Millisecond, 0},
	}

	for _, c := range cases {
		deviceConfig := *basicDeviceConfig
		deviceConfig.WriteBytesPerSecond = c.bytesPerSecond
		deviceConfig.SeekTime = c.seekTime
		writeBackCache := newWriteBackCache(&deviceConfig)
		if got, want := writeBackCache.computeWritableBytes(c.duration), c.want; got != want {
			t.Errorf("computeWritableBytes(%s, %d, %s) = %d, want %d", c.duration, c.bytesPerSecond, c.seekTime, got, want)
		}
	}
}

func TestSliceShuffle(t *testing.T) {
	a := []string{"brown", "dog", "fox", "hello", "jumped", "lazy", "over", "quick", "the", "the", "world"}
	acopy := make([]string, len(a))
	copy(acopy, a)

	sliceShuffle(acopy)
	sort.Strings(acopy)
	if !reflect.DeepEqual(a, acopy) {
		t.Errorf("sliceShuffle failed: %v -> %v", a, acopy)
	}
}
