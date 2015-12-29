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
	"math/rand"
	"slowfs/slowfs"
	"slowfs/slowfs/units"
	"time"
)

type writeBackCache struct {
	// Records cached writes for files. Will be written back gradually or on fsync.
	unwrittenBytes map[string]units.NumBytes

	// If a file is closed while still having writes not yet written back to disk,
	// record them here. If a file is closed we still need to write back data for it, as that
	// will take up spare IO time that would otherwise be used for other files getting written back.
	orphanedUnwrittenBytes units.NumBytes

	deviceConfig *slowfs.DeviceConfig
}

func newWriteBackCache(config *slowfs.DeviceConfig) *writeBackCache {
	return &writeBackCache{
		unwrittenBytes: make(map[string]units.NumBytes),
		deviceConfig:   config,
	}
}

func (wbc *writeBackCache) close(path string) {
	wbc.orphanedUnwrittenBytes += wbc.unwrittenBytes[path]
	delete(wbc.unwrittenBytes, path)
}

func (wbc *writeBackCache) write(path string, numBytes units.NumBytes) {
	if numBytes > 0 {
		wbc.unwrittenBytes[path] += numBytes
	}
}

func (wbc *writeBackCache) getUnwrittenBytes(path string) units.NumBytes {
	return wbc.unwrittenBytes[path]
}

func (wbc *writeBackCache) writeBackFile(path string) {
	delete(wbc.unwrittenBytes, path)
}

func (wbc *writeBackCache) writeBack(duration time.Duration) {
	// Choose random files to write back bytes for.
	paths := make([]string, 0, len(wbc.unwrittenBytes))
	for path := range wbc.unwrittenBytes {
		paths = append(paths, path)
	}

	sliceShuffle(paths)
	for _, path := range paths {
		duration -= wbc.writeBackBytesForFile(path, duration)

		if duration <= 0 {
			break
		}
	}

	if duration >= wbc.deviceConfig.SeekTime {
		wbc.orphanedUnwrittenBytes -= units.NumBytesMin(wbc.orphanedUnwrittenBytes, wbc.computeWritableBytes(duration))
	}

}

func (wbc *writeBackCache) writeBackBytesForFile(path string, duration time.Duration) time.Duration {
	var timeTaken time.Duration
	bytesToWrite := units.NumBytesMin(wbc.unwrittenBytes[path], wbc.computeWritableBytes(duration))

	if bytesToWrite != 0 {
		timeTaken = wbc.deviceConfig.SeekTime + wbc.deviceConfig.WriteTime(bytesToWrite)
	}

	wbc.unwrittenBytes[path] -= bytesToWrite
	if wbc.unwrittenBytes[path] == 0 {
		delete(wbc.unwrittenBytes, path)
	}
	return timeTaken
}

// We assume a seek before we can begin writing back data, so if we don't have time for that seek
// we can't write any bytes back.
func (wbc *writeBackCache) computeWritableBytes(duration time.Duration) units.NumBytes {
	return wbc.deviceConfig.WritableBytes(duration - wbc.deviceConfig.SeekTime)
}

func sliceShuffle(arr []string) {
	for i := 0; i < len(arr); i++ {
		idx := i + rand.Intn(len(arr)-i)
		arr[i], arr[idx] = arr[idx], arr[i]
	}
}
