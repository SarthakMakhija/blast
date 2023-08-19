package tests

import (
	"net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type EchoServer struct {
	listener                     net.Listener
	payloadSizeBytes             int64
	stopChannel                  chan struct{}
	totalRequests                atomic.Uint32
	donotWritebackEveryKRequests uint
}

func NewEchoServer(network, address string, payloadSizeBytes int64) (*EchoServer, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &EchoServer{
		listener:         listener,
		payloadSizeBytes: payloadSizeBytes,
		stopChannel:      make(chan struct{}),
	}, nil
}

func NewEchoServerWithNoWriteback(
	network, address string,
	payloadSizeBytes int64,
	donotWritebackEveryKRequests uint,
) (*EchoServer, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &EchoServer{
		listener:                     listener,
		payloadSizeBytes:             payloadSizeBytes,
		stopChannel:                  make(chan struct{}),
		donotWritebackEveryKRequests: donotWritebackEveryKRequests,
	}, nil
}

func (server *EchoServer) accept(t *testing.T) {
	go func() {
		for {
			select {
			case <-server.stopChannel:
				return
			default:
				connection, err := server.listener.Accept()
				assert.Nil(t, err)

				server.handleConnection(connection)
			}
		}
	}()
}

func (server *EchoServer) handleConnection(connection net.Conn) {
	go func() {
		requestCount := uint(0)
		for {
			select {
			case <-server.stopChannel:
				return
			default:
				payload := make([]byte, server.payloadSizeBytes)

				_, _ = connection.Read(payload)
				server.totalRequests.Add(1)

				requestCount = requestCount + 1
				if server.donotWritebackEveryKRequests != 0 &&
					requestCount%server.donotWritebackEveryKRequests == 0 {
				} else {
					_, _ = connection.Write(payload)
				}
			}
		}
	}()
}

func (server *EchoServer) stop() {
	close(server.stopChannel)
}

func (server *EchoServer) totalRequestsReceived() uint32 {
	return server.totalRequests.Load()
}
