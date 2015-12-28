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

package scheduler

import (
	"math"
	"slowfs/slowfs"
	"time"
)

// ReadWriteQueue reorders any requests that are close enough together in time if they would become
// a sequential read or write by being reordered.
type readWriteQueue struct {
	dc    *deviceContext
	timer *time.Timer
	queue []*requestData
}

func newReadWriteQueue(dc *deviceContext) *readWriteQueue {
	// If we don't stop the timer, it will immediately fire its channel, but we only want to fire
	// when the request at the front of the queue is ready.
	t := time.NewTimer(time.Hour)
	t.Stop()
	return &readWriteQueue{
		dc:    dc,
		timer: t,
		queue: make([]*requestData, 0, 16),
	}
}

func (rwq *readWriteQueue) push(data *requestData) {
	req := data.req
	reqByteEnd := req.Start + req.Size
	var bestDiff slowfs.NumBytes = math.MaxInt64
	bestIdx := len(rwq.queue)
	for i := len(rwq.queue) - 1; i >= 0; i-- {
		otherReq := rwq.queue[i].req

		otherReqByteEnd := otherReq.Start + otherReq.Size
		if otherReq.Path == req.Path && req.Start >= otherReqByteEnd {
			// Place after request other.
			diff := req.Start - otherReqByteEnd
			if diff < bestDiff {
				bestDiff = diff
				bestIdx = i + 1
			}
		}

		// Don't insert before a request that was made really early.
		if req.Timestamp.After(otherReq.Timestamp.Add(rwq.dc.deviceConfig.RequestReorderMaxDelay)) {
			break
		}

		if otherReq.Path == req.Path && reqByteEnd <= otherReq.Start {
			// Place before request other.
			diff := otherReq.Start - reqByteEnd
			if diff < bestDiff {
				bestDiff = diff
				bestIdx = i
			}
		}
	}
	rwq.queue = append(rwq.queue, nil)
	copy(rwq.queue[bestIdx+1:], rwq.queue[bestIdx:])
	rwq.queue[bestIdx] = data
}

func (rwq *readWriteQueue) pop(curTime time.Time) *requestData {
	if len(rwq.queue) == 0 || !rwq.ready(curTime) {
		return nil
	}

	item := rwq.queue[0]
	rwq.queue = rwq.queue[1:]
	return item
}

func (rwq *readWriteQueue) scheduleResponse(curTime time.Time) {
	if len(rwq.queue) == 0 {
		return
	}
	timeToWait := rwq.cutoffTime(rwq.queue[0].req).Sub(curTime)
	rwq.timer.Reset(timeToWait)
}

func (rwq *readWriteQueue) responseChannel() <-chan time.Time {
	return rwq.timer.C
}

func (rwq *readWriteQueue) ready(curTime time.Time) bool {
	if len(rwq.queue) == 0 {
		return false
	}
	return curTime.After(rwq.cutoffTime(rwq.queue[0].req))
}

// We need to wait for a while before allowing a request to be popped off, because requests that
// we may want to put ahead of that request need time to come in. So, we wait half the time that
// the request on the head of the queue takes before saying it can be popped off.
func (rwq *readWriteQueue) cutoffTime(req *Request) time.Time {
	return req.Timestamp.Add(rwq.dc.computeTime(req) / 2)
}
