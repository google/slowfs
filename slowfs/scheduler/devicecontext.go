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
	// Describes the physical medium.
	deviceConfig slowfs.DeviceConfig

	// For the last accessed file, record the offset of the first byte we have not accessed.
	// This is used to determine if reads are sequential or not.
	firstUnseenByte int64

	// Accesses to different files are assumed to be non-sequential reads.
	lastAccessedFile string

	// The device can only execute one request at a time, so record when it is busy until.
	busyUntil time.Time

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
	requestDuration := time.Duration(0)

	switch req.Type {
	case OpenRequest:
		// Leave at 0 seconds.
	case CloseRequest:
		// Leave at 0 seconds.
	case ReadRequest:
		requestDuration = dc.computeSeekTime(req) + computeTimeFromThroughput(req.Size, dc.deviceConfig.ReadBytesPerSecond)
	case WriteRequest:
		// TODO(edcourtney): Implement simulation of write-caching + fsyncs.
		requestDuration = dc.computeSeekTime(req) + computeTimeFromThroughput(req.Size, dc.deviceConfig.WriteBytesPerSecond)
	default:
		dc.logger.Printf("unknown request type for %+v\n", req)
	}

	return latestTime(dc.busyUntil, req.Timestamp).Add(requestDuration).Sub(req.Timestamp)
}

// Execute executes a given request, applying changes to the device context.
func (dc *deviceContext) execute(req *Request) {
	dc.busyUntil = req.Timestamp.Add(dc.computeTime(req))

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

func latestTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func computeTimeFromThroughput(numBytes, bytesPerSecond int64) time.Duration {
	return time.Duration(float64(numBytes) / float64(bytesPerSecond) * float64(time.Second))
}
