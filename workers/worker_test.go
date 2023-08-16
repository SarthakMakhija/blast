package workers

import (
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
		options: WorkerOptions{
			totalRequests:          uint(1),
			payload:                []byte("payload"),
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
		options: WorkerOptions{
			totalRequests:          totalRequests,
			payload:                []byte("payload"),
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
		options: WorkerOptions{
			totalRequests:          totalRequests,
			payload:                []byte("payload"),
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
