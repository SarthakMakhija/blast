package tests

import (
	"net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockServer struct {
	listener         net.Listener
	payloadSizeBytes uint
	stopChannel      chan struct{}
	totalRequests    atomic.Uint32
}

func NewMockServer(network, address string, payloadSizeBytes uint) (*MockServer, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &MockServer{
		listener:         listener,
		payloadSizeBytes: payloadSizeBytes,
		stopChannel:      make(chan struct{}),
	}, nil
}

func (server *MockServer) accept(t *testing.T) {
	go func() {
		connection, err := server.listener.Accept()
		assert.Nil(t, err)

		server.handleConnection(connection)
	}()
}

func (server *MockServer) handleConnection(connection net.Conn) {
	go func() {
		for {
			select {
			case <-server.stopChannel:
				return
			default:
				payload := make([]byte, server.payloadSizeBytes)
				_, _ = connection.Read(payload)
				server.totalRequests.Add(1)
			}
		}
	}()
}

func (server *MockServer) stop() {
	close(server.stopChannel)
}

func (server *MockServer) totalRequestsReceived() uint32 {
	return server.totalRequests.Load()
}
