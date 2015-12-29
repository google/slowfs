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

	if *fsyncStrategy != "" {
		config.FsyncStrategy, err = slowfs.ParseFsyncStrategyFromString(*fsyncStrategy)
		if err != nil {
			log.Fatalf("flag fsync-strategy: %s", err)
		}
	}

	if *writeStrategy != "" {
		config.WriteStrategy, err = slowfs.ParseWriteStrategyFromString(*writeStrategy)
		if err != nil {
			log.Fatalf("flag write-strategy: %s", err)
		}
	}

	if *metadataOpTime != "" {
		config.MetadataOpTime, err = time.ParseDuration(*metadataOpTime)
		if err != nil {
			log.Fatalf("flag metadata-op-time: %s", err)
		}
	}

	if config.WriteStrategy == slowfs.SimulateWrite && config.FsyncStrategy == slowfs.WriteBackCachedFsync {
		log.Printf("setting both simulated writes and write back cache is probably not what you want. " +
			"Write back cache is meant to simulate writes being cached in memory and taking minimal time, " +
			"then being written back to disk later, either during spare IO time or at an fsync.")
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
