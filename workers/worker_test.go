package workers

import (
	"blast/payload"
	"bufio"
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"blast/report"
)

type BytesWriteCloser struct {
	*bufio.Writer
}

func (writeCloser *BytesWriteCloser) Close() error {
	return nil
}

func TestWritesPayloadByWorker(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse, 1)
	defer close(loadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			totalRequests:          uint(1),
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	response := <-worker.options.loadGenerationResponse

	assert.Nil(t, response.Err)
	assert.Equal(t, int64(7), response.PayloadLengthBytes)
}

func TestWritesMultiplePayloadsByWorker(t *testing.T) {
	totalRequests := uint(5)
	loadGenerationResponse := make(chan report.LoadGenerationResponse, totalRequests)
	defer close(loadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			totalRequests:          totalRequests,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	for count := 1; count <= int(totalRequests); count++ {
		response := <-loadGenerationResponse
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(7), response.PayloadLengthBytes)
	}
}

func TestWritesMultiplePayloadsByWorkerWithThrottle(t *testing.T) {
	totalRequests := uint(5)
	loadGenerationResponse := make(chan report.LoadGenerationResponse, totalRequests)
	defer close(loadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection: &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:  NewRequestId(),
		options: WorkerOptions{
			totalRequests:          totalRequests,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
			requestsPerSecond:      float64(3),
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	for count := 1; count <= int(totalRequests); count++ {
		response := <-loadGenerationResponse
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(7), response.PayloadLengthBytes)
	}
}

func TestWritesOnANilConnectionWithConnectionId(t *testing.T) {
	totalRequests := uint(2)
	loadGenerationResponse := make(chan report.LoadGenerationResponse, totalRequests)
	defer close(loadGenerationResponse)

	worker := Worker{
		connection: nil,
		requestId:  NewRequestId(),
		options: WorkerOptions{
			totalRequests:          totalRequests,
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	for count := 1; count <= int(totalRequests); count++ {
		response := <-loadGenerationResponse
		assert.Error(t, response.Err)
		assert.Equal(t, -1, response.ConnectionId)
		assert.Equal(t, ErrNilConnection, response.Err)
	}
}

func TestWritesPayloadByWorkerWithConnectionId(t *testing.T) {
	loadGenerationResponse := make(chan report.LoadGenerationResponse, 1)
	defer close(loadGenerationResponse)

	var buffer bytes.Buffer
	worker := Worker{
		connection:   &BytesWriteCloser{bufio.NewWriter(&buffer)},
		requestId:    NewRequestId(),
		connectionId: 10,
		options: WorkerOptions{
			totalRequests:          uint(1),
			payloadGenerator:       payload.NewConstantPayloadGenerator([]byte("payload")),
			loadGenerationResponse: loadGenerationResponse,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	worker.run(&wg)
	wg.Wait()

	response := <-worker.options.loadGenerationResponse

	assert.Nil(t, response.Err)
	assert.Equal(t, 10, response.ConnectionId)
}
