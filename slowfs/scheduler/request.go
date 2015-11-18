package scheduler

import "time"

type RequestType int64

// Enumeration of different types of requests.
const (
	ReadRequest RequestType = iota
	WriteRequest
	OpenRequest
	CloseRequest
)

// Contains information for all types of requests.
type Request struct {
	Type      RequestType
	Timestamp time.Time
	Path      string
	Start     int64
	Size      int64
}
