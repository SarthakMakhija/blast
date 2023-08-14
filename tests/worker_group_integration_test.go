package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"blast/workers"
)

func TestSendRequests(t *testing.T) {
	server, err := NewMockServer("tcp", "localhost:8080", 10)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	responseChannel := workers.
		NewWorkerGroup(workers.NewGroupOptions(10, 20, []byte("HelloWorld"), "localhost:8080")).
		Run()

	for response := range responseChannel {
		assert.Nil(t, response.Err)
		assert.Equal(t, int64(10), response.PayloadLength)
	}
}
