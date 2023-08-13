package blast

import (
	"io"
	"net"
	"sync"
	"time"
)

const ResponseChannelSize = 10000

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

type WorkerGroup struct {
	options     GroupOptions
	stopChannel chan struct{}
}

type WorkerResponse struct {
	err           error
	payloadLength int64
}

type Worker struct {
	connection io.WriteCloser
	options    WorkerOptions
}

func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return &WorkerGroup{options: options, stopChannel: make(chan struct{}, options.concurrency)}
}

func (group *WorkerGroup) Run() chan WorkerResponse {
	responseChannel := make(chan WorkerResponse, ResponseChannelSize)
	group.runWorkers(responseChannel)
	group.finish(responseChannel)

	return responseChannel
}

func (group *WorkerGroup) runWorkers(responseChannel chan WorkerResponse) {
	var wg sync.WaitGroup
	wg.Add(int(group.options.concurrency))

	connection, err := net.Dial("tcp", group.options.targetAddress)
	if err != nil {
		// TODO: Handle error
		return
	}

	// TODO: Should we care about the responses?
	for count := 1; count <= int(group.options.concurrency); count++ {
		Worker{
			connection: connection,
			options: WorkerOptions{
				totalRequests:     uint(group.options.totalRequests / group.options.concurrency),
				payload:           group.options.payload,
				targetAddress:     group.options.targetAddress,
				requestsPerSecond: group.options.requestsPerSecond,
				stopChannel:       group.stopChannel,
				responseChannel:   responseChannel,
			},
		}.run(&wg)
	}
	wg.Wait()
}

func (group *WorkerGroup) finish(responseChannel chan WorkerResponse) {
	close(responseChannel)
}

func (worker Worker) run(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		worker.sendRequests()
	}()
}

func (worker Worker) sendRequests() {
	var throttle <-chan time.Time
	if worker.options.requestsPerSecond > 0 {
		throttle = time.Tick(
			time.Duration(1e6/(worker.options.requestsPerSecond)) * time.Microsecond,
		)
	}

	for request := 1; request <= int(worker.options.totalRequests); request++ {
		select {
		case <-worker.options.stopChannel:
			return
		default:
			if worker.options.requestsPerSecond > 0 {
				<-throttle
			}
			worker.sendRequest()
		}
	}
}

func (worker Worker) sendRequest() {
	_, err := worker.connection.Write(worker.options.payload)
	worker.options.responseChannel <- WorkerResponse{
		err:           err,
		payloadLength: int64(len(worker.options.payload)),
	}
}
