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

const NilConnectionId int = -1

type LoadGenerationResponse struct {
	Err                error
	PayloadLengthBytes int64
	LoadGenerationTime time.Time
	ConnectionId       int
}

type SubjectServerResponse struct {
	Err                error
	ResponseTime       time.Time
	PayloadLengthBytes int64
}

type ResponseReader struct {
	responseSizeBytes       int64
	readDeadline            time.Duration
	readTotalResponses      atomic.Uint64
	readSuccessfulResponses atomic.Uint64
	stopChannel             chan struct{}
	responseChannel         chan SubjectServerResponse
}

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

func (responseReader *ResponseReader) StartReading(connection net.Conn) {
	go func(connection net.Conn) {
		for {
			defer func() {
				_ = connection.Close()
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "[ResponseReader] %v\n", err.(error).Error())
				}
			}()

			select {
			case <-responseReader.stopChannel:
				return
			default:
				if responseReader.readDeadline != time.Duration(0) {
					connection.SetReadDeadline(time.Now().Add(responseReader.readDeadline))
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

func (responseReader *ResponseReader) Close() {
	close(responseReader.stopChannel)
}

func (responseReader *ResponseReader) TotalResponsesRead() uint64 {
	return responseReader.readTotalResponses.Load()
}

func (responseReader *ResponseReader) TotalSuccessfulResponsesRead() uint64 {
	return responseReader.readSuccessfulResponses.Load()
}
