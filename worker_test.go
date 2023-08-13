package blast

import (
	"bufio"
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type BytesWriteCloser struct {
	*bufio.Writer
}

func (writeCloser *BytesWriteCloser) Close() error {
	return nil
}

func TestWritesPayloadByWorker(t *testing.T) {
	responseChannel := make(chan WorkerResponse, 1)
	defer close(responseChannel)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		options: WorkerOptions{
			totalRequests:   uint(1),
			payload:         []byte("payload"),
			responseChannel: responseChannel,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	response := <-worker.options.responseChannel

	assert.Nil(t, response.err)
	assert.Equal(t, int64(7), response.payloadLength)
}

func TestWritesMultiplePayloadsByWorker(t *testing.T) {
	totalRequests := uint(5)
	responseChannel := make(chan WorkerResponse, totalRequests)
	defer close(responseChannel)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		options: WorkerOptions{
			totalRequests:   totalRequests,
			payload:         []byte("payload"),
			responseChannel: responseChannel,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	for count := 1; count <= int(totalRequests); count++ {
		response := <-responseChannel
		assert.Nil(t, response.err)
		assert.Equal(t, int64(7), response.payloadLength)
	}
}

func TestWritesMultiplePayloadsByWorkerWithThrottle(t *testing.T) {
	totalRequests := uint(5)
	responseChannel := make(chan WorkerResponse, totalRequests)
	defer close(responseChannel)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		options: WorkerOptions{
			totalRequests:     totalRequests,
			payload:           []byte("payload"),
			responseChannel:   responseChannel,
			requestsPerSecond: float64(3),
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	for count := 1; count <= int(totalRequests); count++ {
		response := <-responseChannel
		assert.Nil(t, response.err)
		assert.Equal(t, int64(7), response.payloadLength)
	}
}
