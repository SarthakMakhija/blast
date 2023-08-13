package blast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendRequests(t *testing.T) {
	server, err := NewMockServer("tcp", "localhost:8080", 10)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	responseChannel := NewWorkerGroup(GroupOptions{
		concurrency:   10,
		totalRequests: 20,
		payload:       []byte("HelloWorld"),
		targetAddress: "localhost:8080",
	}).Run()

	for response := range responseChannel {
		assert.Nil(t, response.err)
		assert.Equal(t, int64(10), response.payloadLength)
	}
}
