// Package scheduler provides the scheduler, which takes requests of particular types
// (e.g. read, write, open) and returns how long they should wait before completing.
package scheduler

import (
	"log"
	"slowfs/slowfs"
	"time"
)

// Scheduler determines how long operations should take given a description of a physical medium.
type Scheduler struct {
	// Describes the physical media.
	deviceConfig slowfs.DeviceConfig

	requests chan *requestData
}

// New creates a new Scheduler using the given DeviceConfig to help compute how long requests
// should take.
func New(config slowfs.DeviceConfig) *Scheduler {
	scheduler := &Scheduler{
		deviceConfig: config,
		requests:     make(chan *requestData, 10),
	}
	go scheduler.serveRequests()
	return scheduler
}

type requestData struct {
	req             *Request
	responseChannel chan time.Duration
}

// Schedule schedules a new request and returns how long the request should take.
// N.B. this can block.
func (s *Scheduler) Schedule(req *Request) time.Duration {
	ch := make(chan time.Duration, 1)
	s.requests <- &requestData{req, ch}
	return <-ch
}

// Main event loop to serve requests.
func (s *Scheduler) serveRequests() {
	for {
		reqData := <-s.requests
		req, resp := reqData.req, reqData.responseChannel
		reqDuration := time.Duration(0)

		switch req.Type {
		case OpenRequest:
			reqDuration = 100 * time.Millisecond
		case CloseRequest:
			reqDuration = 100 * time.Millisecond
		case ReadRequest:
			reqDuration = 10 * time.Millisecond
		case WriteRequest:
			reqDuration = 5 * time.Millisecond
		default:
			log.Printf("unknown request type: %v\n", req)
		}

		resp <- reqDuration
	}
}
