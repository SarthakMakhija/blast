package workers

import (
	"net"
	"sync"
)

const ResponseChannelSize = 10000

type WorkerGroup struct {
	options     GroupOptions
	stopChannel chan struct{}
}

func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return &WorkerGroup{options: options, stopChannel: make(chan struct{}, options.concurrency)}
}

func (group *WorkerGroup) Run() chan LoadGenerationResponse {
	loadGenerationResponse := make(chan LoadGenerationResponse, ResponseChannelSize)
	group.runWorkers(loadGenerationResponse)
	group.finish(loadGenerationResponse)

	return loadGenerationResponse
}

// TODO: Should we care about the responses?
func (group *WorkerGroup) runWorkers(loadGenerationResponse chan LoadGenerationResponse) {
	var wg sync.WaitGroup
	wg.Add(int(group.options.concurrency))

	connectionsSharedByWorker := group.options.concurrency / group.options.connections

	var connection net.Conn
	var err error

	for count := 0; count < int(group.options.concurrency); count++ {
		if count%int(connectionsSharedByWorker) == 0 {
			connection, err = net.Dial("tcp", group.options.targetAddress)
			if err != nil {
				// TODO: Handle error
				return
			}
		}

		Worker{
			connection: connection,
			options: WorkerOptions{
				totalRequests: uint(
					group.options.totalRequests / group.options.concurrency,
				),
				payload:                group.options.payload,
				targetAddress:          group.options.targetAddress,
				requestsPerSecond:      group.options.requestsPerSecond,
				stopChannel:            group.stopChannel,
				loadGenerationResponse: loadGenerationResponse,
			},
		}.run(&wg)
	}
	wg.Wait()
}

func (group *WorkerGroup) finish(loadGenerationResponse chan LoadGenerationResponse) {
	close(loadGenerationResponse)
}
