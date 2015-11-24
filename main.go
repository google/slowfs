package main

import (
	"flag"
	"log"
	"path/filepath"
	"slowfs/slowfs"
	"slowfs/slowfs/fuselayer"
	"slowfs/slowfs/scheduler"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func main() {
	backingDir := flag.String("backing-dir", "", "directory to use as storage")
	mountDir := flag.String("mount-dir", "", "directory to mount at")
	fsyncStrategy := flag.String("fsync-strategy", "writebackcache", "choice of none/no, dumb, writebackcache/wbc")
	flag.Parse()

	if *backingDir == "" || *mountDir == "" {
		log.Fatalf("arguments backing-dir and mount-dir are required.")
	}

	var err error

	*backingDir, err = filepath.Abs(*backingDir)
	if err != nil {
		log.Fatalf("invalid backing-dir: %v", err)
	}

	*mountDir, err = filepath.Abs(*mountDir)
	if err != nil {
		log.Fatalf("invalid mount-dir: %v", err)
	}

	if *backingDir == *mountDir {
		log.Fatalf("backing directory may not be the same as mount directory.")
	}

	config := slowfs.HardDriveDeviceConfig

	switch *fsyncStrategy {
	case "none", "no":
		config.FsyncStrategy = slowfs.NoFsync
	case "dumb":
		config.FsyncStrategy = slowfs.DumbFsync
	case "writebackcache", "wbc":
		config.FsyncStrategy = slowfs.WriteBackCachedFsync
	default:
		log.Fatalf("unknown fsync strategy %s.", *fsyncStrategy)
	}

	scheduler := scheduler.New(&config)
	fs := pathfs.NewPathNodeFs(fuselayer.NewSlowFs(*backingDir, scheduler), nil)
	server, _, err := nodefs.MountRoot(*mountDir, fs.Root(), nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	server.Serve()
}
