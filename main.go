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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"slowfs/slowfs"
	"slowfs/slowfs/fuselayer"
	"slowfs/slowfs/scheduler"
	"time"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func main() {
	configs := map[string]*slowfs.DeviceConfig{
		slowfs.HDD7200RpmDeviceConfig.Name: &slowfs.HDD7200RpmDeviceConfig,
	}

	backingDir := flag.String("backing-dir", "", "directory to use as storage")
	mountDir := flag.String("mount-dir", "", "directory to mount at")

	configFile := flag.String("config-file", "", "path to config file listing device configurations")
	configName := flag.String("config-name", "hdd7200rpm", "which config to use (built-ins: hdd7200rpm)")

	// Flags for overriding any subset of the config. These are all strings (even the durations)
	// because we need to differentiate between the flag not being specified, and being set to the
	// default value.
	seekWindow := flag.String("seek-window", "", "")
	seekTime := flag.String("seek-time", "", "")
	readBytesPerSecond := flag.String("read-bytes-per-second", "", "")
	writeBytesPerSecond := flag.String("write-bytes-per-second", "", "")
	allocateBytesPerSecond := flag.String("allocate-bytes-per-second", "", "")
	requestReorderMaxDelay := flag.String("request-reorder-max-delay", "", "")
	fsyncStrategy := flag.String("fsync-strategy", "", "choice of none/no, dumb, writebackcache/wbc")
	writeStrategy := flag.String("write-strategy", "", "choice of fast, simulate")
	metadataOpTime := flag.String("metadata-op-time", "", "duration value (e.g. 10ms)")
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

	if *configFile != "" {
		data, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("couldn't read config file %s: %s", *configFile, err)
		}
		dcs, err := slowfs.ParseDeviceConfigsFromJSON(data)
		if err != nil {
			log.Fatalf("couldn't parse config file %s: %s", *configFile, err)
		}
		for _, dc := range dcs {
			if _, ok := configs[dc.Name]; ok {
				log.Fatalf("duplicate device config with name '%s'", dc.Name)
			}
			configs[dc.Name] = dc
		}
	}

	config, ok := configs[*configName]

	if !ok {
		log.Fatalf("unknown config %s", *configName)
	}

	flagsHadError := false

	if *seekWindow != "" {
		config.SeekWindow, err = slowfs.ParseNumBytesFromString(*seekWindow)
		if err != nil {
			log.Printf("flag seek-window: %s", err)
			flagsHadError = true
		}
	}

	if *seekTime != "" {
		config.SeekTime, err = time.ParseDuration(*seekTime)
		if err != nil {
			log.Printf("flag seek-time: %s", err)
			flagsHadError = true
		}
	}

	if *readBytesPerSecond != "" {
		config.ReadBytesPerSecond, err = slowfs.ParseNumBytesFromString(*readBytesPerSecond)
		if err != nil {
			log.Printf("flag read-bytes-per-second: %s", err)
			flagsHadError = true
		}
	}

	if *writeBytesPerSecond != "" {
		config.WriteBytesPerSecond, err = slowfs.ParseNumBytesFromString(*writeBytesPerSecond)
		if err != nil {
			log.Printf("flag write-bytes-per-second: %s", err)
			flagsHadError = true
		}
	}

	if *allocateBytesPerSecond != "" {
		config.AllocateBytesPerSecond, err = slowfs.ParseNumBytesFromString(*allocateBytesPerSecond)
		if err != nil {
			log.Printf("flag allocate-bytes-per-second: %s", err)
			flagsHadError = true
		}
	}

	if *requestReorderMaxDelay != "" {
		config.RequestReorderMaxDelay, err = time.ParseDuration(*requestReorderMaxDelay)
		if err != nil {
			log.Printf("flag request-reorder-max-delay: %s", err)
			flagsHadError = true
		}
	}

	if *fsyncStrategy != "" {
		config.FsyncStrategy, err = slowfs.ParseFsyncStrategyFromString(*fsyncStrategy)
		if err != nil {
			log.Printf("flag fsync-strategy: %s", err)
			flagsHadError = true
		}
	}

	if *writeStrategy != "" {
		config.WriteStrategy, err = slowfs.ParseWriteStrategyFromString(*writeStrategy)
		if err != nil {
			log.Printf("flag write-strategy: %s", err)
			flagsHadError = true
		}
	}

	if *metadataOpTime != "" {
		config.MetadataOpTime, err = time.ParseDuration(*metadataOpTime)
		if err != nil {
			log.Printf("flag metadata-op-time: %s", err)
			flagsHadError = true
		}
	}

	if flagsHadError {
		log.Fatalf("flags had error(s), exiting")
	}

	err = config.Validate()
	if err != nil {
		log.Fatalf("error validating config: %s", err)
	}

	fmt.Printf("using config: %s\n", config)
	scheduler := scheduler.New(config)
	fs := pathfs.NewPathNodeFs(fuselayer.NewSlowFs(*backingDir, scheduler), nil)
	server, _, err := nodefs.MountRoot(*mountDir, fs.Root(), nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	server.Serve()
}
