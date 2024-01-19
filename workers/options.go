package workers

import (
	"blast/payload"
	"time"

	"blast/report"
)

const dialTimeout = 3 * time.Second

// GroupOptions defines the configuration options for the WorkerGroup.
type GroupOptions struct {
	concurrency       uint
	connections       uint
	totalRequests     uint
	payloadGenerator  payload.PayloadGenerator
	targetAddress     string
	requestsPerSecond float64
	dialTimeout       time.Duration
}

// WorkerOptions defines the configuration options for a running Worker.
type WorkerOptions struct {
	totalRequests          uint
	payloadGenerator       payload.PayloadGenerator
	targetAddress          string
	requestsPerSecond      float64
	stopChannel            chan struct{}
	loadGenerationResponse chan report.LoadGenerationResponse
}

// NewGroupOptions creates a new instance of GroupOptions.
func NewGroupOptions(
	concurrency uint,
	totalRequests uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		totalRequests,
		payloadGenerator,
		targetAddress,
		0.0,
		dialTimeout,
	)
}

// NewGroupOptionsWithConnections creates a new instance of GroupOptions.
func NewGroupOptionsWithConnections(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		connections,
		totalRequests,
		payloadGenerator,
		targetAddress,
		0.0,
		dialTimeout,
	)
}

// NewGroupOptionsFullyLoaded creates a new instance of GroupOptions.
func NewGroupOptionsFullyLoaded(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payloadGenerator payload.PayloadGenerator,
	targetAddress string,
	requestsPerSecond float64,
	dialTimeout time.Duration,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		connections:       connections,
		totalRequests:     totalRequests,
		payloadGenerator:  payloadGenerator,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
		dialTimeout:       dialTimeout,
	}
}

// TotalRequests returns the total number of requests set in GroupOptions.
func (groupOptions GroupOptions) TotalRequests() uint {
	return groupOptions.totalRequests
}
