package report

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type LoadGenerationResponse struct {
	Err                error
	PayloadLengthBytes int64
	LoadGenerationTime time.Time
}

type SubjectServerResponse struct {
	Err                error
	ResponseTime       time.Time
	PayloadLengthBytes int64
}

type ResponseReader struct {
	responseSizeBytes    int64
	totalResponsesToRead uint32
	readResponses        atomic.Uint32
	stopChannel          chan struct{}
	responseChannel      chan SubjectServerResponse
}

func NewResponseReader(
	responseSizeBytes int64,
	totalResponsesToRead uint,
	responseChannel chan SubjectServerResponse,
) *ResponseReader {
	return &ResponseReader{
		responseSizeBytes:    responseSizeBytes,
		totalResponsesToRead: uint32(totalResponsesToRead),
		stopChannel:          make(chan struct{}),
		responseChannel:      responseChannel,
	}
}

func (responseReader *ResponseReader) StartReading(connection net.Conn) {
	go func(connection net.Conn) {
		for {
			defer func() {
				if err := recover(); err != nil {
					println("received error in ResponseReader", err.(error).Error())
				}
			}()

			select {
			case <-responseReader.stopChannel:
				return
			default:
				connection.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				buffer := make([]byte, responseReader.responseSizeBytes)
				_, err := connection.Read(buffer)

				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
					responseReader.responseChannel <- SubjectServerResponse{
						Err:          err,
						ResponseTime: time.Now(),
					}
				} else {
					responseReader.responseChannel <- SubjectServerResponse{
						ResponseTime:       time.Now(),
						PayloadLengthBytes: int64(len(buffer)),
					}
				}
				responseReader.readResponses.Add(1)
			}
		}
	}(connection)
}

func (responseReader *ResponseReader) close() {
	close(responseReader.stopChannel)
}

func (responseReader *ResponseReader) TotalResponsesRead() uint32 {
	return responseReader.readResponses.Load()
}
