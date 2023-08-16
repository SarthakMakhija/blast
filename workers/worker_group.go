package workers

import (
	"net"
	"sync"

	"blast/report"
)

type WorkerGroup struct {
	options     GroupOptions
	stopChannel chan struct{}
}

func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return &WorkerGroup{options: options, stopChannel: make(chan struct{}, options.concurrency)}
}

func (group *WorkerGroup) Run() chan report.LoadGenerationResponse {
	loadGenerationResponse := make(chan report.LoadGenerationResponse, group.options.totalRequests)
	group.runWorkers(loadGenerationResponse)
	group.finish(loadGenerationResponse)

	return loadGenerationResponse
}

func (group *WorkerGroup) runWorkers(loadGenerationResponse chan report.LoadGenerationResponse) {
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

func (group *WorkerGroup) finish(loadGenerationResponse chan report.LoadGenerationResponse) {
	close(loadGenerationResponse)
}
