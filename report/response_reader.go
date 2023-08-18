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
	responseSizeBytes       int64
	readTotalResponses      atomic.Uint64
	readSuccessfulResponses atomic.Uint64
	stopChannel             chan struct{}
	responseChannel         chan SubjectServerResponse
}

func NewResponseReader(
	responseSizeBytes int64,
	responseChannel chan SubjectServerResponse,
) *ResponseReader {
	return &ResponseReader{
		responseSizeBytes: responseSizeBytes,
		stopChannel:       make(chan struct{}), // TODO: size of the channel?
		responseChannel:   responseChannel,
	}
}

func (responseReader *ResponseReader) StartReading(connection net.Conn) {
	go func(connection net.Conn) {
		for {
			defer func() {
				_ = connection.Close()
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
					responseReader.readSuccessfulResponses.Add(1)
					responseReader.responseChannel <- SubjectServerResponse{
						ResponseTime:       time.Now(),
						PayloadLengthBytes: int64(len(buffer)),
					}
				}
				responseReader.readTotalResponses.Add(1)
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
