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
	"slowfs/slowfs"
	"testing"
	"time"
)

func TestLatestTime(t *testing.T) {
	cases := []struct {
		a    time.Time
		b    time.Time
		want time.Time
	}{
		{startTime, startTime, startTime},
		{startTime, startTime.Add(time.Millisecond), startTime.Add(time.Millisecond)},
		{startTime.Add(time.Millisecond), startTime, startTime.Add(time.Millisecond)},
		{startTime.Add(-5 * time.Microsecond), startTime, startTime},
	}

	for _, c := range cases {
		if got, want := latestTime(c.a, c.b), c.want; got != want {
			t.Errorf("latestTime(%s, %s) = %s, want %s", c.a, c.b, got, want)
		}
	}
}

func TestDeviceContext_ComputeTimeAndExecute(t *testing.T) {
	type requestInvocation struct {
		req  *Request
		want time.Duration
	}

	cases := []struct {
		desc         string
		deviceConfig *slowfs.DeviceConfig
		requests     []requestInvocation
	}{
		{
			desc:         "sequential read",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(110 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 100 * time.Millisecond,
				},
			},
		},
		{
			desc:         "sequential write",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 10 * time.Millisecond,
				},
			},
		},
		{
			desc:         "backwards read",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(110 * time.Millisecond),
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
			},
		},
		{
			desc:         "backwards write",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
			},
		},
		{
			desc:         "spaced out read",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(110 * time.Millisecond),
						Path:      "a",
						Start:     5,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(220 * time.Millisecond),
						Path:      "a",
						Start:     8,
						Size:      1,
					},
					want: 100 * time.Millisecond,
				},
			},
		},
		{
			desc:         "spaced out write",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "a",
						Start:     5,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(40 * time.Millisecond),
						Path:      "a",
						Start:     8,
						Size:      1,
					},
					want: 10 * time.Millisecond,
				},
			},
		},
		{
			desc:         "multiple files read",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime.Add(110 * time.Millisecond),
						Path:      "b",
						Start:     1,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
			},
		},
		{
			desc:         "multiple files write",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(20 * time.Millisecond),
						Path:      "b",
						Start:     1,
						Size:      1,
					},
					want: 20 * time.Millisecond,
				},
			},
		},
		{
			desc:         "close",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      CloseRequest,
						Timestamp: startTime,
						Path:      "a",
					},
					want: 80 * time.Millisecond,
				},
			},
		},
		{
			desc:         "device busy",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 110 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      ReadRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 210 * time.Millisecond,
				},
			},
		},
		{
			desc:         "metadata",
			deviceConfig: readWriteAsymmetricDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      MetadataRequest,
						Timestamp: startTime,
					},
					want: 80 * time.Millisecond,
				},
			},
		},
		{
			desc:         "fast writes",
			deviceConfig: fastWriteDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1,
					},
					want: 0 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime.Add(10 * time.Millisecond),
						Path:      "a",
						Start:     1,
						Size:      1,
					},
					want: 0 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     10,
						Size:      100,
					},
					want: 10 * time.Millisecond, // Busy until previous req finishes.
				},
			},
		},
		{
			desc:         "write back cache",
			deviceConfig: writeBackCacheDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     0,
						Size:      1000,
					},
					want: 0 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      WriteRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     1000,
						Size:      100,
					},
					want: 0 * time.Millisecond,
				},
				{
					req: &Request{
						Type:      FsyncRequest,
						Timestamp: startTime,
						Path:      "a",
					},
					want: 11*time.Second + 10*time.Millisecond,
				},
			},
		},
		{
			desc:         "allocate",
			deviceConfig: basicDeviceConfig,
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      AllocateRequest,
						Timestamp: startTime,
						Path:      "a",
						Start:     5,
						Size:      4123,
					},
					want: 4123*time.Millisecond + 10*time.Millisecond,
				},
			},
		},
	}

	for _, c := range cases {
		dc := newDeviceContext(c.deviceConfig)
		for _, req := range c.requests {
			if got, want := dc.computeTime(req.req), req.want; got != want {
				t.Errorf("fail (%s) computeTime(%+v) = %s, want %s", c.desc, req.req, got, want)
			}
			dc.execute(req.req)
		}
	}
}
