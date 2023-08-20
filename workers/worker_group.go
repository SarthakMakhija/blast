package workers

import (
	"fmt"
	"net"
	"os"
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

func (group *WorkerGroup) runWorkers(
	loadGenerationResponseChannel chan report.LoadGenerationResponse,
) {
	var wg sync.WaitGroup
	wg.Add(int(group.options.concurrency))

	connectionsSharedByWorker := group.options.concurrency / group.options.connections

	var connection net.Conn
	var err error

	var connectionId int = -1
	for count := 0; count < int(group.options.concurrency); count++ {
		if count%int(connectionsSharedByWorker) == 0 || connection == nil {
			connection, err = group.newConnection()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[WorkerGroup] %v\n", err.Error())
			} else {
				connectionId = connectionId + 1
			}
			if group.responseReader != nil && connection != nil {
				group.responseReader.StartReading(connection)
			}
		}
		group.runWorker(connection, connectionId, &wg, loadGenerationResponseChannel)
	}
	wg.Wait()
	group.doneChannel <- struct{}{}
}

func (group *WorkerGroup) WaitTillDone() {
	<-group.doneChannel
}

func (group *WorkerGroup) DoneChannel() chan struct{} {
	return group.doneChannel
}

func (group *WorkerGroup) newConnection() (net.Conn, error) {
	connection, err := net.DialTimeout(
		"tcp",
		group.options.targetAddress,
		group.options.dialTimeout,
	)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (group *WorkerGroup) runWorker(
	connection net.Conn,
	connectionId int,
	wg *sync.WaitGroup,
	loadGenerationResponseChannel chan report.LoadGenerationResponse,
) {
	totalRequests := group.options.totalRequests
	Worker{
		connection:   connection,
		connectionId: connectionId,
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
