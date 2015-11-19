package scheduler

import (
	"slowfs/slowfs"
	"testing"
	"time"
)

var startTime time.Time

var testDeviceConfig = slowfs.DeviceConfig{
	SeekWindow:          4 * slowfs.Byte,
	SeekTime:            10 * time.Millisecond,
	ReadBytesPerSecond:  10 * slowfs.Byte,
	WriteBytesPerSecond: 100 * slowfs.Byte,
}

func TestComputeTimeFromThroughput(t *testing.T) {
	cases := []struct {
		numBytes       int64
		bytesPerSecond int64
		duration       time.Duration
	}{
		{1, 1, 1 * time.Second},
		{0, 1, 0 * time.Second},
		{1, 1000, 1 * time.Millisecond},
		{1000, 1, 1000 * time.Second},
		{3, 9, 333333333 * time.Nanosecond},
	}

	for _, c := range cases {
		if got, want := computeTimeFromThroughput(c.numBytes, c.bytesPerSecond), c.duration; got != want {
			t.Errorf("computeTimeFromThroughput(%d, %d) = %s, want %s",
				c.numBytes, c.bytesPerSecond, got, want)
		}
	}

}

func TestDeviceContext_ComputeTimeAndExecute(t *testing.T) {
	type requestInvocation struct {
		req  *Request
		want time.Duration
	}

	cases := []struct {
		desc     string
		requests []requestInvocation
	}{
		{
			desc: "sequential read",
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
			desc: "sequential write",
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
			desc: "backwards read",
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
			desc: "backwards write",
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
			desc: "spaced out read",
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
			desc: "spaced out write",
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
			desc: "multiple files read",
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
			desc: "multiple files write",
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
			desc: "open",
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      OpenRequest,
						Timestamp: startTime,
						Path:      "a",
					},
					want: 0 * time.Millisecond,
				},
			},
		},
		{
			desc: "close",
			requests: []requestInvocation{
				{
					req: &Request{
						Type:      CloseRequest,
						Timestamp: startTime,
						Path:      "a",
					},
					want: 0 * time.Millisecond,
				},
			},
		},
	}

	for _, c := range cases {
		dc := newDeviceContext(testDeviceConfig)
		for _, req := range c.requests {
			if got, want := dc.computeTime(req.req), req.want; got != want {
				t.Errorf("fail (%s) computeTime(%+v) = %s, want %s", c.desc, req.req, got, want)
			}
			dc.execute(req.req)
		}
	}
}
