package workers

import (
	"time"

	"blast/report"
)

const dialTimeout = 3 * time.Second

type GroupOptions struct {
	concurrency       uint
	connections       uint
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
	dialTimeout       time.Duration
}

type WorkerOptions struct {
	totalRequests          uint
	payload                []byte
	targetAddress          string
	requestsPerSecond      float64
	stopChannel            chan struct{}
	loadGenerationResponse chan report.LoadGenerationResponse
}

func NewGroupOptions(
	concurrency uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		1,
		totalRequests,
		payload,
		targetAddress,
		0.0,
		dialTimeout,
	)
}

func NewGroupOptionsWithConnections(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsFullyLoaded(
		concurrency,
		connections,
		totalRequests,
		payload,
		targetAddress,
		0.0,
		dialTimeout,
	)
}

func NewGroupOptionsFullyLoaded(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
	requestsPerSecond float64,
	dialTimeout time.Duration,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		connections:       connections,
		totalRequests:     totalRequests,
		payload:           payload,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
		dialTimeout:       dialTimeout,
	}
}

func (groupOptions GroupOptions) TotalRequests() uint {
	return groupOptions.totalRequests
}
