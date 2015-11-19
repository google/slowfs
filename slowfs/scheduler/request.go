package scheduler

import "time"

// RequestType denotes what type a request is.
type RequestType int64

// Enumeration of different types of requests.
const (
	ReadRequest RequestType = iota
	WriteRequest
	OpenRequest
	CloseRequest
)

// Request contains information for all types of requests.
type Request struct {
	Type      RequestType
	Timestamp time.Time
	Path      string
	Start     int64
	Size      int64
}
