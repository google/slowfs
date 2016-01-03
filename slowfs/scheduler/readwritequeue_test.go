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
	"fmt"
	"reflect"
	"testing"
	"time"
)

func (rd *requestData) String() string {
	return fmt.Sprintf("req: %v chan: %v", rd.req, rd.responseChannel)
}

func TestReadWriteQueue_CutoffTime(t *testing.T) {
	var startTime time.Time
	testRwq := newReadWriteQueue(newDeviceContext(basicDeviceConfig))

	cases := []struct {
		desc string
		req  *Request
		want time.Time
	}{
		{
			desc: "20 ms request",
			req: &Request{
				Type:      ReadRequest,
				Timestamp: startTime,
				Path:      "a",
				Start:     0,
				Size:      1,
			},
			want: startTime.Add(10 * time.Millisecond),
		},
	}

	for _, c := range cases {
		if got, want := testRwq.cutoffTime(c.req), c.want; got != want {
			t.Errorf("fail (%s) cutoffTime(%+v) = %s, want %s", c.desc, c.req, got, want)
		}
	}
}

func TestReadWriteQueue_Ready(t *testing.T) {
	var startTime time.Time

	cases := []struct {
		desc       string
		reqData    []*requestData
		falseTimes []time.Time
		trueTimes  []time.Time
	}{
		{
			desc: "no request",
			falseTimes: []time.Time{
				startTime,
				startTime.Add(4 * time.Millisecond),
				startTime.Add(100 * time.Millisecond),
			},
		},
		{
			desc: "20 ms request",
			reqData: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
			},
			falseTimes: []time.Time{
				startTime,
				startTime.Add(4 * time.Millisecond),
				startTime.Add(10 * time.Millisecond),
			},
			trueTimes: []time.Time{
				startTime.Add(11 * time.Millisecond),
				startTime.Add(100 * time.Millisecond),
			},
		},
		{
			desc: "re-ordered requests",
			reqData: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(5 * time.Millisecond),
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
			},
			falseTimes: []time.Time{
				startTime,
				startTime.Add(10 * time.Millisecond),
				startTime.Add(15 * time.Millisecond),
			},
			trueTimes: []time.Time{
				startTime.Add(16 * time.Millisecond),
				startTime.Add(100 * time.Millisecond),
			},
		},
	}

	for _, c := range cases {
		var testRwq = newReadWriteQueue(newDeviceContext(basicDeviceConfig))
		for _, reqData := range c.reqData {
			testRwq.push(reqData)
		}

		for _, falseTime := range c.falseTimes {
			if got, want := testRwq.ready(falseTime), false; got != want {
				t.Errorf("fail (%s) ready(%s) = %t, want %t", c.desc, falseTime, got, want)
			}
		}
		for _, trueTime := range c.trueTimes {
			if got, want := testRwq.ready(trueTime), true; got != want {
				t.Errorf("fail (%s) ready(%s) = %t, want %t", c.desc, trueTime, got, want)
			}
		}

	}
}

func TestReadWriteQueue_ScheduleResponse(t *testing.T) {
	var startTime time.Time

	cases := []struct {
		desc    string
		reqData *requestData
		curTime time.Time
		want    time.Duration
	}{
		{
			desc: "20 ms request, curTime: 0 ms",
			reqData: &requestData{
				&Request{
					Type:      ReadRequest,
					Timestamp: startTime,
					Path:      "a",
					Start:     0,
					Size:      1,
				},
				nil,
			},
			curTime: startTime,
			want:    10 * time.Millisecond,
		},
		{
			desc: "20 ms request, curTime: 5 ms",
			reqData: &requestData{
				&Request{
					Type:      ReadRequest,
					Timestamp: startTime,
					Path:      "a",
					Start:     0,
					Size:      1,
				},
				nil,
			},
			curTime: startTime.Add(5 * time.Millisecond),
			want:    5 * time.Millisecond,
		},
		{
			desc: "20 ms request, curTime: 25 ms",
			reqData: &requestData{
				&Request{
					Type:      ReadRequest,
					Timestamp: startTime,
					Path:      "a",
					Start:     0,
					Size:      1,
				},
				nil,
			},
			curTime: startTime.Add(25 * time.Millisecond),
			want:    0 * time.Millisecond,
		},
	}

	for _, c := range cases {
		var testRwq = newReadWriteQueue(newDeviceContext(basicDeviceConfig))
		testRwq.push(c.reqData)
		testRwq.scheduleResponse(c.curTime)
		start := time.Now()
		<-testRwq.responseChannel()
		if got, want := time.Since(start), c.want; got-want < 0 || got-want > time.Millisecond {
			t.Errorf("fail (%s) response took %s, want %s", c.desc, got, want)
		}
	}
}

func TestReadWriteQueue_PushAndPop(t *testing.T) {
	var startTime time.Time

	type popInvocation struct {
		time time.Time
		want *requestData
	}

	cases := []struct {
		desc   string
		pushes []*requestData
		pops   []popInvocation
	}{
		{
			desc: "nothing",
			pops: []popInvocation{
				{
					startTime.Add(10 * time.Millisecond),
					nil,
				},
			},
		},
		{
			desc: "20 ms request",
			pushes: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
			},
			pops: []popInvocation{
				{
					startTime,
					nil,
				},
				{
					startTime.Add(11 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "a",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(20 * time.Millisecond),
					nil,
				},
			},
		},
		{
			desc: "request starving",
			pushes: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "b",
						Start:     0,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(70 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					nil,
				},
			},
			pops: []popInvocation{
				{
					startTime,
					nil,
				},
				{
					startTime.Add(11 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "a",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(30 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(31 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(20 * time.Millisecond),
							Path:      "b",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(80 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(81 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(70 * time.Millisecond),
							Path:      "a",
							Start:     1,
							Size:      1,
						},
						nil,
					},
				},
			},
		},
		{
			desc: "request starving 2",
			pushes: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(4 * time.Millisecond),
						Path:      "b",
						Start:     0,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(5 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(10 * time.Millisecond),
						Path:      "a",
						Start:     2,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(15 * time.Millisecond),
						Path:      "a",
						Start:     3,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "a",
						Start:     4,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(25 * time.Millisecond),
						Path:      "a",
						Start:     5,
						Size:      1,
					},
					nil,
				},
			},
			pops: []popInvocation{
				{
					startTime.Add(10 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(11 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "a",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(17 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(18 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(5 * time.Millisecond),
							Path:      "a",
							Start:     1,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(25 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(26 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(10 * time.Millisecond),
							Path:      "a",
							Start:     2,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(32 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(33 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(4 * time.Millisecond),
							Path:      "b",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(47 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(48 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(15 * time.Millisecond),
							Path:      "a",
							Start:     3,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(55 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(56 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(20 * time.Millisecond),
							Path:      "a",
							Start:     4,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(62 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(63 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(25 * time.Millisecond),
							Path:      "a",
							Start:     5,
							Size:      1,
						},
						nil,
					},
				},
			},
		},
		{
			desc: "big test",
			pushes: []*requestData{
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(3 * time.Millisecond),
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(1 * time.Millisecond),
						Path:      "a",
						Start:     20,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "b",
						Start:     21,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     11,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     40,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(1 * time.Millisecond),
						Path:      "a",
						Start:     30,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(1 * time.Millisecond),
						Path:      "a",
						Start:     2,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(11 * time.Millisecond),
						Path:      "a",
						Start:     10,
						Size:      1,
					},
					nil,
				},
				{
					&Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(2 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					nil,
				},
			},
			pops: []popInvocation{
				{
					startTime.Add(13 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(14 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(3 * time.Millisecond),
							Path:      "a",
							Start:     0,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(17 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(18 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(2 * time.Millisecond),
							Path:      "a",
							Start:     1,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(22 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(23 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(1 * time.Millisecond),
							Path:      "a",
							Start:     2,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(31 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(32 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "a",
							Start:     11,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(42 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(43 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(1 * time.Millisecond),
							Path:      "a",
							Start:     20,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(52 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(53 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(1 * time.Millisecond),
							Path:      "a",
							Start:     30,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(61 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(62 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "a",
							Start:     40,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(71 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(72 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime,
							Path:      "b",
							Start:     21,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(87 * time.Millisecond),
					nil,
				},
				{
					startTime.Add(88 * time.Millisecond),
					&requestData{
						&Request{
							Type:      ReadRequest,
							Timestamp: startTime.Add(11 * time.Millisecond),
							Path:      "a",
							Start:     10,
							Size:      1,
						},
						nil,
					},
				},
				{
					startTime.Add(500 * time.Millisecond),
					nil,
				},
			},
		},
	}

	for _, c := range cases {
		dc := newDeviceContext(basicDeviceConfig)
		testRwq := newReadWriteQueue(dc)
		for _, push := range c.pushes {
			testRwq.push(push)
		}
		for _, pop := range c.pops {
			got := testRwq.pop(pop.time)
			if !reflect.DeepEqual(got, pop.want) {
				t.Errorf("fail (%s) pop(%+v) = %+v, want %+v", c.desc, pop.time, got, pop.want)
			}
			if pop.want != nil {
				dc.execute(pop.want.req)
			}
		}
	}
}
