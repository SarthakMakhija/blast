package workers

import (
	"net"
	"sync"

	"blast/report"
)

type WorkerGroup struct {
	options        GroupOptions
	stopChannel    chan struct{}
	doneChannel    chan struct{}
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
		doneChannel:    make(chan struct{}, 1),
		responseReader: responseReader,
	}
}

func (group *WorkerGroup) Run() chan report.LoadGenerationResponse {
	if group.options.totalRequests%group.options.concurrency != 0 {
		group.options.totalRequests = ((group.options.totalRequests / group.options.concurrency) + 1) * group.options.concurrency
	}
	loadGenerationResponseChannel := make(
		chan report.LoadGenerationResponse,
		group.options.totalRequests,
	)

	go func() {
		group.runWorkers(loadGenerationResponseChannel)
		group.WaitTillDone()
		return
	}()
	return loadGenerationResponseChannel
}

func (group *WorkerGroup) Close() {
	for count := 1; count <= int(group.options.concurrency); count++ {
		group.stopChannel <- struct{}{}
	}
}

// TODO: close the connection if response reader is nil
func (group *WorkerGroup) runWorkers(
	loadGenerationResponseChannel chan report.LoadGenerationResponse,
) {
	var wg sync.WaitGroup
	wg.Add(int(group.options.concurrency))

	connectionsSharedByWorker := group.options.concurrency / group.options.connections

	var connection net.Conn
	for count := 0; count < int(group.options.concurrency); count++ {
		if count%int(connectionsSharedByWorker) == 0 {
			connection, _ = group.newConnection()
			// TODO: Handle error
			if group.responseReader != nil && connection != nil {
				group.responseReader.StartReading(connection)
			}
		}
		group.runWorker(connection, &wg, loadGenerationResponseChannel)
	}
	wg.Wait()
	group.doneChannel <- struct{}{}
}

func (group *WorkerGroup) WaitTillDone() {
	<-group.doneChannel
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
	totalRequests := group.options.totalRequests
	Worker{
		connection: connection,
		options: WorkerOptions{
			totalRequests: uint(
				totalRequests / group.options.concurrency,
			),
			payload:                group.options.payload,
			targetAddress:          group.options.targetAddress,
			requestsPerSecond:      group.options.requestsPerSecond,
			stopChannel:            group.stopChannel,
			loadGenerationResponse: loadGenerationResponseChannel,
		},
	}.run(wg)
}
