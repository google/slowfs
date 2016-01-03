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

// Package fuselayer contains the go-fuse handling code.
package fuselayer

import (
	"slowfs/slowfs/scheduler"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type slowFile struct {
	nodefs.File

	path string
	sfs  *SlowFs
}

// Read performs a read, and then waits until the scheduled time.
func (sf *slowFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	start := time.Now()
	r, status := sf.File.Read(dest, off)
	// TODO(edcourtney): How long should it take in the case of an error?
	if status != fuse.OK {
		return r, status
	}

	// The read doesn't actually get executed until we do it explicitly, so do it now.
	// If we don't, time will get spent doing the read where we don't expect.
	buf := make([]byte, r.Size())
	buf, status = r.Bytes(buf)
	// TODO(edcourtney): How long should it take in the case of an error?
	if status != fuse.OK {
		return nil, status
	}
	r = fuse.ReadResultData(buf)

	opTime := sf.sfs.scheduler.Schedule(&scheduler.Request{
		Type:      scheduler.ReadRequest,
		Timestamp: start,
		Path:      sf.path,
		Start:     off,
		Size:      int64(r.Size()),
	})

	time.Sleep(opTime - time.Since(start))

	return r, status
}

// Write performs a write, and then waits until the scheduled time.
func (sf *slowFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	start := time.Now()
	// Unlike Read, Write will immediately execute the syscall.
	r, status := sf.File.Write(data, off)

	// TODO(edcourtney): How long should it take in the case of an error?
	if status != fuse.OK {
		return r, status
	}

	opTime := sf.sfs.scheduler.Schedule(&scheduler.Request{
		Type:      scheduler.WriteRequest,
		Timestamp: start,
		Path:      sf.path,
		Start:     off,
		Size:      int64(r),
	})

	time.Sleep(opTime - time.Since(start))

	return r, status
}

// Release calls Release on the underlying file, and then waits until the scheduled time.
func (sf *slowFile) Release() {
	start := time.Now()
	sf.File.Release()

	opTime := sf.sfs.scheduler.Schedule(&scheduler.Request{
		Type:      scheduler.CloseRequest,
		Timestamp: start,
		Path:      sf.path,
	})
	time.Sleep(opTime - time.Since(start))
}

func (sf *slowFile) Fsync(flags int) fuse.Status {
	start := time.Now()
	r := sf.File.Fsync(flags)
	// TODO(edcourtney): How long should this take?
	if r != fuse.OK {
		return r
	}

	opTime := sf.sfs.scheduler.Schedule(&scheduler.Request{
		Type:      scheduler.FsyncRequest,
		Timestamp: start,
		Path:      sf.path,
	})
	time.Sleep(opTime - time.Since(start))

	return r
}

// SlowFs is a FileSystem whose operations take amounts of time determined by an associated
// Scheduler.
type SlowFs struct {
	pathfs.FileSystem

	scheduler *scheduler.Scheduler
}

// NewSlowFs creates a new SlowFs using the specified scheduler at the given directory. The
// directory must be empty.
func NewSlowFs(directory string, scheduler *scheduler.Scheduler) *SlowFs {
	return &SlowFs{
		FileSystem: pathfs.NewLoopbackFileSystem(directory),
		scheduler:  scheduler,
	}
}

// Open opens a file, and then waits until the scheduled time.
func (sfs *SlowFs) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	start := time.Now()
	file, status := sfs.FileSystem.Open(name, flags, context)
	// TODO(edcourtney): How long should it take in the case of an error?
	if status != fuse.OK {
		return file, status
	}

	slowFile := &slowFile{
		File: file,
		sfs:  sfs,
		path: name,
	}

	opTime := sfs.scheduler.Schedule(&scheduler.Request{
		Type:      scheduler.OpenRequest,
		Timestamp: start,
		Path:      slowFile.path,
	})
	time.Sleep(opTime - time.Since(start))

	return slowFile, status
}
