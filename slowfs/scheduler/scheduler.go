// Package scheduler provides the scheduler, which takes requests of particular types
// (e.g. read, write, open) and returns how long they should wait before completing.
package scheduler

import (
	"slowfs/slowfs"
	"time"
)

// Scheduler determines how long operations should take given a description of a physical medium.
type Scheduler struct {
	dc             *deviceContext
	readWriteQueue *readWriteQueue
	requests       chan *requestData
}

// New creates a new Scheduler using the given DeviceConfig to help compute how long requests
// should take.
func New(config *slowfs.DeviceConfig) *Scheduler {
	dc := newDeviceContext(config)
	scheduler := &Scheduler{
		dc:             dc,
		readWriteQueue: newReadWriteQueue(dc),
		requests:       make(chan *requestData, 10),
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
		select {
		case reqData := <-s.requests:
			req, resp := reqData.req, reqData.responseChannel
			switch req.Type {
			case ReadRequest, WriteRequest:
				s.readWriteQueue.push(reqData)
			default:
				resp <- s.dc.computeTime(req)
				s.dc.execute(req)
			}
		case <-s.readWriteQueue.responseChannel():
			reqData := s.readWriteQueue.pop(time.Now())
			if reqData != nil {
				reqData.responseChannel <- s.dc.computeTime(reqData.req)
				s.dc.execute(reqData.req)
			}
		}

		// This needs to be called every loop, since executing a request can change how long a
		// read or write request on the front of the queue would take.
		s.readWriteQueue.scheduleResponse(time.Now())
	}
}
