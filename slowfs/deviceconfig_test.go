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
	"testing"
	"time"
)

func TestComputeTimeFromThroughput(t *testing.T) {
	cases := []struct {
		numBytes       int64
		bytesPerSecond int64
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
		bytesPerSecond int64
		want           int64
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
