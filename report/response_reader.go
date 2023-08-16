package report

import (
	"errors"
	"io"
	"net"
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
	Response           []byte
	PayloadLengthBytes int64
}

type ResponseReader struct {
	responseSizeBytes uint
	stopChannel       chan struct{}
	responseChannel   chan SubjectServerResponse
}

func NewResponseReader(
	responseSizeBytes uint,
	stopChannel chan struct{},
	responseChannel chan SubjectServerResponse,
) *ResponseReader {
	return &ResponseReader{
		responseSizeBytes: responseSizeBytes,
		stopChannel:       stopChannel,
		responseChannel:   responseChannel,
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
						Response:           buffer,
						PayloadLengthBytes: int64(len(buffer)),
					}
				}
			}
		}
	}(connection)
}

func (responseReader *ResponseReader) close() {
	close(responseReader.stopChannel)
}
