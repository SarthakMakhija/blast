package workers

import (
	"blast/report"
)

type GroupOptions struct {
	concurrency       uint
	connections       uint
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
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
	)
}

func NewGroupOptionsFullyLoaded(
	concurrency uint,
	connections uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
	requestsPerSecond float64,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		connections:       connections,
		totalRequests:     totalRequests,
		payload:           payload,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
	}
}

func (groupOptions GroupOptions) TotalRequests() uint {
	return groupOptions.totalRequests
}
