package report

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// NilConnectionId represents the connection id for an unestablished connection.
const NilConnectionId int = -1

// LoadGenerationResponse represents the load generated on the target server.
type LoadGenerationResponse struct {
	Err                error
	PayloadLengthBytes int64
	LoadGenerationTime time.Time
	ConnectionId       int
}

// SubjectServerResponse represents the response read from the target server.
type SubjectServerResponse struct {
	Err                error
	ResponseTime       time.Time
	PayloadLengthBytes int64
}

// ResponseReader reads the response from the specified net.Conn.
type ResponseReader struct {
	responseSizeBytes       int64
	readDeadline            time.Duration
	readTotalResponses      atomic.Uint64
	readSuccessfulResponses atomic.Uint64
	stopChannel             chan struct{}
	responseChannel         chan SubjectServerResponse
}

// NewResponseReader creates a new instance of ResponseReader.
// All the read responses are sent to responseChannel.
func NewResponseReader(
	responseSizeBytes int64,
	readDeadline time.Duration,
	responseChannel chan SubjectServerResponse,
) *ResponseReader {
	return &ResponseReader{
		responseSizeBytes: responseSizeBytes,
		readDeadline:      readDeadline,
		stopChannel:       make(chan struct{}),
		responseChannel:   responseChannel,
	}
}

// StartReading runs a goroutine that reads from the provided net.Conn.
// It keeps on reading from the connection until either of the two happen:
// 1) Reading from the connection returns an io.EOF error
// 2) ResponseReader gets stopped
// ResponseReader implements one goroutine for each new connection created by the workers.WorkerGroup.
func (responseReader *ResponseReader) StartReading(connection net.Conn) {
	go func(connection net.Conn) {
		defer func() {
			_ = connection.Close()
			if err := recover(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "[ResponseReader] %v\n", err.(error).Error())
			}
		}()

		for {
			select {
			case <-responseReader.stopChannel:
				return
			default:
				if responseReader.readDeadline != time.Duration(0) {
					_ = connection.SetReadDeadline(time.Now().Add(responseReader.readDeadline))
				}
				buffer := make([]byte, responseReader.responseSizeBytes)
				n, err := connection.Read(buffer)

				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
					responseReader.readTotalResponses.Add(1)
					responseReader.responseChannel <- SubjectServerResponse{
						Err:          err,
						ResponseTime: time.Now(),
					}
				} else if n > 0 && buffer != nil && len(buffer) > 0 {
					responseReader.readSuccessfulResponses.Add(1)
					responseReader.readTotalResponses.Add(1)
					responseReader.responseChannel <- SubjectServerResponse{
						ResponseTime:       time.Now(),
						PayloadLengthBytes: int64(len(buffer)),
					}
				}
			}
		}
	}(connection)
}

// Close closes the stopChannel which stops all the goroutines.
func (responseReader *ResponseReader) Close() {
	close(responseReader.stopChannel)
}

// TotalResponsesRead returns the total responses read from the target server.
// It includes successful and failed responses.
func (responseReader *ResponseReader) TotalResponsesRead() uint64 {
	return responseReader.readTotalResponses.Load()
}

// TotalSuccessfulResponsesRead returns the total successful responses read from the target server.
func (responseReader *ResponseReader) TotalSuccessfulResponsesRead() uint64 {
	return responseReader.readSuccessfulResponses.Load()
}
