// Contains the go-fuse handling code.
package fuselayer

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type SlowFile struct {
	nodefs.File

	sfs *SlowFs
}

func NewSlowFile(file nodefs.File, sfs *SlowFs) *SlowFile {
	return &SlowFile{
		File: file,
		sfs:  sfs,
	}
}

func (sf *SlowFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	r, status := sf.File.Read(dest, off)
	if status != fuse.OK {
		return r, status
	}

	// The read doesn't actually get executed until we do it explicitly, so do it now.
	// If we don't, time will get spent doing the read where we don't expect.
	buf := make([]byte, r.Size())
	buf, status = r.Bytes(buf)
	if status != fuse.OK {
		return nil, status
	}
	r = fuse.ReadResultData(buf)

	return r, status
}

func (sf *SlowFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	r, status := sf.File.Write(data, off)

	return r, status
}

func (sf *SlowFile) Release() {
	sf.File.Release()
}

type SlowFs struct {
	pathfs.FileSystem
}

func NewSlowFs(directory string) *SlowFs {
	return &SlowFs{
		FileSystem: pathfs.NewLoopbackFileSystem(directory),
	}
}

func (sfs *SlowFs) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	file, status := sfs.FileSystem.Open(name, flags, context)
	if status != fuse.OK {
		return file, status
	}

	return NewSlowFile(file, sfs), status
}
