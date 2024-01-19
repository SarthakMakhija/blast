package workers

import (
	"errors"
	"io"
	"sync"
	"time"

	"blast/report"
)

// ErrNilConnection is the error that is returned when the worker operates on an unestablished
// connection. This can happen if the connection to the target server can not be established.
var ErrNilConnection = errors.New("attempting to send request on a nil connection")

// Worker sends load on the target connection.
// connection field is usually a net.Conn.
// Each connection is also given a unique connection id that is used for reporting.
type Worker struct {
	connection   io.WriteCloser
	connectionId int
	options      WorkerOptions
	requestId    *RequestId
}

// run runs a Worker.
// The total number of requests that a Worker sends is equal to the totalRequests/concurrency.
func (worker Worker) run(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		worker.sendRequests()
	}()
}

// sendRequests sends worker.options.totalRequests on the connection.
// Each Worker can be stopped by closing the stopChannel, even before the totalRequests are sent.
// Each worker also implements throttle if worker.options.requestsPerSecond > 0.
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

// sendRequest sends a single request.
// sendRequest sends a single request.
// The result of sending the request is sent on the channel identified by worker.options.loadGenerationResponse.
func (worker Worker) sendRequest() {
	defer func() {
		_ = recover()
	}()
	if worker.connection != nil {
		payload := worker.options.payloadGenerator.Generate(worker.requestId.Next())
		_, err := worker.connection.Write(payload)

		worker.options.loadGenerationResponse <- report.LoadGenerationResponse{
			Err:                err,
			PayloadLengthBytes: int64(len(payload)),
			LoadGenerationTime: time.Now(),
			ConnectionId:       worker.connectionId,
		}
		return
	}
	worker.options.loadGenerationResponse <- report.LoadGenerationResponse{
		Err:                ErrNilConnection,
		PayloadLengthBytes: 0,
		LoadGenerationTime: time.Now(),
		ConnectionId:       report.NilConnectionId,
	}
}
