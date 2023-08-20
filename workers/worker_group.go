package workers

import (
	"fmt"
	"net"
	"os"
	"sync"

	"blast/report"
)

// WorkerGroup is a collection of workers that sends totalRequests to the server.
// WorkerGroup creates a total of options.concurrency Workers.
// Each Worker takes a part of the totalRequests.
// Consider that 100 requests are to be sent with 5 workers, then each Worker will send
// a total of 20 requests.
// Consider that 100 requests are to be sent with 6 workers, then the system will end up
// sending a total of 102 requests, and each Worker will send 17 requests.
// WorkerGroup also provides support for triggering response reading from the connection.
type WorkerGroup struct {
	options        GroupOptions
	stopChannel    chan struct{}
	doneChannel    chan struct{}
	responseReader *report.ResponseReader
}

// NewWorkerGroup returns a new instance of WorkerGroup without supporting reading from the
// connection.
func NewWorkerGroup(options GroupOptions) *WorkerGroup {
	return NewWorkerGroupWithResponseReader(options, nil)
}

// NewWorkerGroupWithResponseReader returns a new instance of WorkerGroup
// that also supports reading from the connection.
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

// Run runs the WorkerGroup and returns a channel of type report.LoadGenerationResponse.
// report.LoadGenerationResponse will contain each request sent by the Worker.
// This method runs a separate goroutine that runs the workers and the goroutine waits until
// all the workers are done.
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

// Closes sends a stop signal to all the workers.
func (group *WorkerGroup) Close() {
	for count := 1; count <= int(group.options.concurrency); count++ {
		group.stopChannel <- struct{}{}
	}
}

// runWorkers runs all the workers.
// The numbers of workers that will run is determined by the concurrency field in GroupOptions.
// These workers will share the tcp connections and the sharing of tcp connections is determined
// by the number of workers and the connections.
// Consider that 100 workers are supposed to be running and blast needs to create 25 connections.
// This configuration will end up sharing a single connection with four workers.
// runWorkers also starts the report.ResponseReader to read from the connection,
// if it is configured to do so.
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

// WaitTillDone waits till all the workers are done.
func (group *WorkerGroup) WaitTillDone() {
	<-group.doneChannel
}

// DoneChannel returns the doneChannel.
func (group *WorkerGroup) DoneChannel() chan struct{} {
	return group.doneChannel
}

// newConnection creates a new TCP connection.
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

// runWorker runs a Worker.
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
