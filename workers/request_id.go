package workers

import "sync/atomic"

// RequestId generates a unique request id for each request.
type RequestId struct {
	next atomic.Uint64
}

// NewRequestId creates a new instance of RequestId.
func NewRequestId() *RequestId {
	return &RequestId{}
}

// Next creates new request id
func (requestId *RequestId) Next() uint64 {
	return requestId.next.Add(1)
}
