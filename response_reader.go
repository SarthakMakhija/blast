package blast

import (
	"errors"
	"io"
	"net"
	"time"
)

type SubjectServerResponse struct {
	Err          error
	ResponseTime time.Time
	Response     []byte
}

type ResponseReader struct {
	responseSizeBytes uint
	connections       []net.Conn
	stopChannel       chan struct{}
	responseChannel   chan SubjectServerResponse
}

func NewResponseReader(
	responseSizeBytes uint,
	connections []net.Conn,
	stopChannel chan struct{},
	responseChannel chan SubjectServerResponse,
) ResponseReader {
	return ResponseReader{
		responseSizeBytes: responseSizeBytes,
		connections:       connections,
		stopChannel:       stopChannel,
		responseChannel:   responseChannel,
	}
}

func (responseReader ResponseReader) StartReading() {
	for index := 0; index < len(responseReader.connections); index++ {
		go func(connection net.Conn) {
			for {
				select {
				case <-responseReader.stopChannel:
					return
				default:
					connection.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
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
							ResponseTime: time.Now(),
							Response:     buffer,
						}
					}
				}
			}
		}(responseReader.connections[index])
	}
}

func (responseReader ResponseReader) close() {
	close(responseReader.stopChannel)
}
