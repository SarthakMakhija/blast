package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"blast/workers"
)

func TestSendRequestsWithSingleConnection(t *testing.T) {
	payloadSizeBytes := uint(10)
	server, err := NewMockServer("tcp", "localhost:8080", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)
	responseChannel := workers.
		NewWorkerGroup(
			workers.NewGroupOptions(
				concurrency,
				totalRequests,
				[]byte("HelloWorld"),
				"localhost:8080",
			),
		).
		Run()

	for response := range responseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLength)
	}
}

func TestSendRequestsWithMultipleConnections(t *testing.T) {
	payloadSizeBytes := uint(10)
	server, err := NewMockServer("tcp", "localhost:8081", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, connections, totalRequests := uint(20), uint(10), uint(40)
	responseChannel := workers.
		NewWorkerGroup(
			workers.NewGroupOptionsWithConnections(
				concurrency,
				connections,
				totalRequests,
				[]byte("HelloWorld"),
				"localhost:8081",
			),
		).
		Run()

	for response := range responseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLength)
	}
}
