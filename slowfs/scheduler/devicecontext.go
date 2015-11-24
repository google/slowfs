package scheduler

import (
	"log"
	"os"
	"slowfs/slowfs"
	"time"
)

// DeviceContext holds the state of the device to determine how long a request should take, taking
// into account things like seeking and sequentiality. This is after any re-ordering has been
// applied. Conceptually this is the actual physical medium -- executing a request here affects
// the state of the device. In this model, we assume that the underlying medium can only run one
// request at a time.
type deviceContext struct {
	// Describes the physical media.
	deviceConfig *slowfs.DeviceConfig

	// For the last accessed file, record the offset of the first byte we have not accessed.
	// This is used to determine if reads are sequential or not.
	firstUnseenByte int64

	// Accesses to different files are assumed to be non-sequential reads.
	lastAccessedFile string

	// The device can only execute one request at a time, so record when it is busy until.
	busyUntil time.Time

	logger *log.Logger

	// Holds information about data not yet written back to disk.
	writeBackCache *writeBackCache
}

// NewDeviceContext creates a new context given a DeviceConfig. DeviceContext will use that
// configuration to compute how long requests take.
func newDeviceContext(config *slowfs.DeviceConfig) *deviceContext {
	var writeBackCache *writeBackCache
	if config.FsyncStrategy == slowfs.WriteBackCachedFsync {
		writeBackCache = newWriteBackCache(config)
	}
	return &deviceContext{
		deviceConfig:   config,
		logger:         log.New(os.Stderr, "DeviceContext: ", log.Ldate|log.Ltime|log.Lshortfile),
		writeBackCache: writeBackCache,
	}
}

// ComputeTime computes how long a request should take given the current state of the device.
// It does not update the context.
func (dc *deviceContext) computeTime(req *Request) time.Duration {
	requestDuration := time.Duration(0)

	switch req.Type {
	case OpenRequest:
		// Leave at 0 seconds.
	case CloseRequest:
		// Leave at 0 seconds.
	case ReadRequest:
		requestDuration = dc.computeSeekTime(req) + dc.deviceConfig.ReadTime(req.Size)
	case WriteRequest:
		switch dc.deviceConfig.WriteStrategy {
		case slowfs.FastWrite:
			// Leave at 0 seconds.
		case slowfs.SimulateWrite:
			requestDuration = dc.computeSeekTime(req) + dc.deviceConfig.WriteTime(req.Size)
		}
	case FsyncRequest:
		switch dc.deviceConfig.FsyncStrategy {
		case slowfs.DumbFsync:
			requestDuration = dc.deviceConfig.SeekTime * 10
		case slowfs.WriteBackCachedFsync:
			requestDuration = dc.deviceConfig.SeekTime + dc.deviceConfig.WriteTime(dc.writeBackCache.getUnwrittenBytes(req.Path))
		}
	default:
		dc.logger.Printf("unknown request type for %+v\n", req)
	}

	return latestTime(dc.busyUntil, req.Timestamp).Add(requestDuration).Sub(req.Timestamp)
}

// Execute executes a given request, applying changes to the device context.
func (dc *deviceContext) execute(req *Request) {
	spareTime := req.Timestamp.Sub(dc.busyUntil)

	// Devote spare time to writing back cache.
	if spareTime > 0 && dc.writeBackCache != nil {
		dc.writeBackCache.writeBack(spareTime)
	}

	dc.busyUntil = req.Timestamp.Add(dc.computeTime(req))

	switch req.Type {
	case OpenRequest:
	case CloseRequest:
		if dc.writeBackCache != nil {
			dc.writeBackCache.close(req.Path)
		}
		if dc.lastAccessedFile == req.Path {
			dc.lastAccessedFile = ""
			dc.firstUnseenByte = 0
		}
	case ReadRequest:
		dc.lastAccessedFile = req.Path
		dc.firstUnseenByte = req.Start + req.Size
	case WriteRequest:
		switch dc.deviceConfig.WriteStrategy {
		case slowfs.FastWrite:
			// Fast writes don't affect things here.
		case slowfs.SimulateWrite:
			dc.lastAccessedFile = req.Path
			dc.firstUnseenByte = req.Start + req.Size
		}

		if dc.writeBackCache != nil {
			dc.writeBackCache.write(req.Path, req.Size)
		}
	case FsyncRequest:
		if dc.writeBackCache != nil {
			dc.writeBackCache.writeBackFile(req.Path)
		}
	default:
		dc.logger.Printf("unknown request type for %+v\n", req)
	}
}

func (dc *deviceContext) computeSeekTime(req *Request) time.Duration {
	// Seek if:
	//   1. We're accessing a different file or an unseen one.
	//   2. We're looking very far ahead compared to last access.
	//   3. We're going backwards.
	if dc.lastAccessedFile != req.Path || dc.firstUnseenByte > req.Start ||
		req.Start-dc.firstUnseenByte >= dc.deviceConfig.SeekWindow {
		return dc.deviceConfig.SeekTime
	}
	return time.Duration(0)
}

func latestTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
