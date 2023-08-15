package tests

import (
	"bytes"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"blast/report"
)

func TestReadsResponseFromASingleConnection(t *testing.T) {
	payloadSizeBytes := uint(10)
	server, err := NewMockServer("tcp", "localhost:9090", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	connection := connectTo(t, "localhost:9090")
	writeTo(t, connection, []byte("HelloWorld"))

	stopChannel := make(chan struct{})
	responseChannel := make(chan report.SubjectServerResponse, 1)

	defer close(stopChannel)
	defer close(responseChannel)

	responseReader := report.NewResponseReader(
		payloadSizeBytes,
		[]net.Conn{connection},
		stopChannel,
		responseChannel,
	)
	responseReader.StartReading()

	response := <-responseChannel

	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("HelloWorld"), response.Response)
}

func TestReadsResponseFromTwoConnections(t *testing.T) {
	payloadSizeBytes := uint(10)
	server, err := NewMockServer("tcp", "localhost:9091", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	connection, otherConnection := connectTo(t, "localhost:9091"), connectTo(t, "localhost:9091")
	writeTo(t, connection, []byte("HelloWorld"))
	writeTo(t, connection, []byte("BlastWorld"))

	time.Sleep(10 * time.Millisecond)

	stopChannel := make(chan struct{})
	responseChannel := make(chan report.SubjectServerResponse, 2)

	defer close(stopChannel)
	defer close(responseChannel)

	responseReader := report.NewResponseReader(
		payloadSizeBytes,
		[]net.Conn{connection, otherConnection},
		stopChannel,
		responseChannel,
	)
	responseReader.StartReading()

	responses := captureTwoResponses(t, responseChannel)
	assert.Equal(t, []byte("BlastWorld"), responses[0])
	assert.Equal(t, []byte("HelloWorld"), responses[1])
}

func connectTo(t *testing.T, address string) net.Conn {
	connection, err := net.Dial("tcp", address)
	assert.Nil(t, err)

	return connection
}

func writeTo(t *testing.T, connection net.Conn, payload []byte) {
	_, err := connection.Write(payload)
	assert.Nil(t, err)
}

func captureTwoResponses(t *testing.T, responseChannel chan report.SubjectServerResponse) [][]byte {
	var responses [][]byte

	response := <-responseChannel
	assert.Nil(t, response.Err)
	responses = append(responses, response.Response)

	response = <-responseChannel
	assert.Nil(t, response.Err)
	responses = append(responses, response.Response)

	sort.Slice(responses, func(i, j int) bool {
		return bytes.Compare(responses[i], responses[j]) <= 0
	})

	return responses
}
