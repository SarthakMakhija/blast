package workers

import (
	"net"
	"sync"

	"blast/report"
)

type WorkerGroup struct {
	options        GroupOptions
	stopChannel    chan struct{}
	responseReader *report.ResponseReader
}

func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return NewWorkerGroupWithResponseReader(options, nil)
}

func NewWorkerGroupWithResponseReader(
	options GroupOptions,
	responseReader *report.ResponseReader,
) *WorkerGroup {
	return &WorkerGroup{
		options:        options,
		stopChannel:    make(chan struct{}, options.concurrency),
		responseReader: responseReader,
	}
}

func (group *WorkerGroup) Run() chan report.LoadGenerationResponse {
	loadGenerationResponseChannel := make(
		chan report.LoadGenerationResponse,
		group.options.totalRequests,
	)
	group.runWorkers(loadGenerationResponseChannel)
	group.finish(loadGenerationResponseChannel)

	return loadGenerationResponseChannel
}

func (group *WorkerGroup) runWorkers(
	loadGenerationResponseChannel chan report.LoadGenerationResponse,
) {
	var wg sync.WaitGroup
	wg.Add(int(group.options.concurrency))

	connectionsSharedByWorker := group.options.concurrency / group.options.connections

	var connection net.Conn
	var err error
	for count := 0; count < int(group.options.concurrency); count++ {
		if count%int(connectionsSharedByWorker) == 0 {
			connection, err = group.newConnection()
			if err != nil {
				// TODO: Handle error
				return
			}
			if group.responseReader != nil {
				group.responseReader.StartReading(connection)
			}
		}
		group.runWorker(connection, &wg, loadGenerationResponseChannel)
	}
	wg.Wait()
}

func (group *WorkerGroup) finish(loadGenerationResponseChannel chan report.LoadGenerationResponse) {
	close(loadGenerationResponseChannel)
}

func (group *WorkerGroup) newConnection() (net.Conn, error) {
	connection, err := net.Dial("tcp", group.options.targetAddress)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (group *WorkerGroup) runWorker(
	connection net.Conn,
	wg *sync.WaitGroup,
	loadGenerationResponseChannel chan report.LoadGenerationResponse,
) {
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
			loadGenerationResponse: loadGenerationResponseChannel,
		},
	}.run(wg)
}
