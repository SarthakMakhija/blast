package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"blast/report"
	"blast/workers"
)

func TestSendsRequestsWithSingleConnection(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:8080", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)
	loadGenerationResponseChannel := workers.
		NewWorkerGroup(
			workers.NewGroupOptions(
				concurrency,
				totalRequests,
				[]byte("HelloWorld"),
				"localhost:8080",
			),
		).Run()

	for response := range loadGenerationResponseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLengthBytes)
	}
}

func TestSendsRequestsWithMultipleConnections(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:8081", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, connections, totalRequests := uint(20), uint(10), uint(40)
	loadGenerationResponseChannel := workers.
		NewWorkerGroup(
			workers.NewGroupOptionsWithConnections(
				concurrency,
				connections,
				totalRequests,
				[]byte("HelloWorld"),
				"localhost:8081",
			),
		).Run()

	for response := range loadGenerationResponseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLengthBytes)
	}
}

func TestSendsARequestAndReadsResponseWithSingleConnection(t *testing.T) {
	payloadSizeBytes, responseSizeBytes := int64(10), int64(10)
	server, err := NewEchoServer("tcp", "localhost:8082", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)

	concurrency, totalRequests := uint(10), uint(20)
	responseChannel := make(chan report.SubjectServerResponse)

	defer func() {
		server.stop()
		close(responseChannel)
	}()

	loadGenerationResponseChannel := workers.NewWorkerGroupWithResponseReader(
		workers.NewGroupOptions(
			concurrency,
			totalRequests,
			[]byte("HelloWorld"),
			"localhost:8082",
		),
		report.NewResponseReader(
			responseSizeBytes,
			responseChannel,
		),
	).Run()

	for response := range loadGenerationResponseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLengthBytes)
	}

	for count := 1; count < int(totalRequests); count++ {
		response := <-responseChannel
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLengthBytes)
	}
}
