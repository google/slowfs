package scheduler

import (
	"log"
	"os"
	"slowfs/slowfs"
	"time"
)

// DeviceContext holds the state of the device to determine how long a request should take.
type deviceContext struct {
	// Describes the physical media.
	deviceConfig slowfs.DeviceConfig

	// For the last accessed file, record the offset of the first byte we have not accessed.
	// This is used to determine if reads are sequential or not.
	firstUnseenByte int64

	// Accesses to different files are assumed to be non-sequential reads.
	lastAccessedFile string

	logger *log.Logger
}

// NewDeviceContext creates a new context given a DeviceConfig. DeviceContext will use that
// configuration to compute how long requests take.
func newDeviceContext(config slowfs.DeviceConfig) *deviceContext {
	return &deviceContext{
		deviceConfig: config,
		logger:       log.New(os.Stderr, "DeviceContext: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// ComputeTime computes how long a request should take given the current state of the device.
// It does not update the context.
func (dc *deviceContext) computeTime(req *Request) time.Duration {
	switch req.Type {
	case OpenRequest:
		return time.Duration(0)
	case CloseRequest:
		return time.Duration(0)
	case ReadRequest:
		return dc.computeSeekTime(req) + computeTimeFromThroughput(req.Size, dc.deviceConfig.ReadBytesPerSecond)
	case WriteRequest:
		// TODO(edcourtney): Implement simulation of write-caching + fsyncs.
		return dc.computeSeekTime(req) + computeTimeFromThroughput(req.Size, dc.deviceConfig.WriteBytesPerSecond)
	default:
		dc.logger.Printf("unknown request type for %+v\n", req)
		return time.Duration(0)
	}
}

// Execute executes a given request, applying changes to the device context.
func (dc *deviceContext) execute(req *Request) {
	switch req.Type {
	case OpenRequest:
	case CloseRequest:
		if dc.lastAccessedFile == req.Path {
			dc.lastAccessedFile = ""
			dc.firstUnseenByte = 0
		}
	case ReadRequest, WriteRequest:
		dc.lastAccessedFile = req.Path
		dc.firstUnseenByte = req.Start + req.Size
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

func computeTimeFromThroughput(numBytes, bytesPerSecond int64) time.Duration {
	return time.Duration(float64(numBytes) / float64(bytesPerSecond) * float64(time.Second))
}
