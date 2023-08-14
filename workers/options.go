package workers

type GroupOptions struct {
	concurrency       uint
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
}

type WorkerOptions struct {
	totalRequests     uint
	payload           []byte
	targetAddress     string
	requestsPerSecond float64
	stopChannel       chan struct{}
	responseChannel   chan WorkerResponse
}

type WorkerResponse struct {
	Err           error
	PayloadLength int64
}

func NewGroupOptions(
	concurrency uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
) GroupOptions {
	return NewGroupOptionsWithRequestsPerSecond(
		concurrency,
		totalRequests,
		payload,
		targetAddress,
		0.0,
	)
}

func NewGroupOptionsWithRequestsPerSecond(
	concurrency uint,
	totalRequests uint,
	payload []byte,
	targetAddress string,
	requestsPerSecond float64,
) GroupOptions {
	return GroupOptions{
		concurrency:       concurrency,
		totalRequests:     totalRequests,
		payload:           payload,
		targetAddress:     targetAddress,
		requestsPerSecond: requestsPerSecond,
	}
}
