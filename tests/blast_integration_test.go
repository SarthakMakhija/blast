package tests

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"blast/blast"
	"blast/workers"
)

func TestBlastWithLoadGeneration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10001", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(10), uint(20)

	groupOptions := workers.NewGroupOptions(
		concurrency,
		totalRequests,
		[]byte("HelloWorld"),
		"localhost:10001",
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer
	blast.NewBlastWithoutResponseReading(groupOptions, 5*time.Minute)

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests: 20"))
	assert.True(t, strings.Contains(output, "SuccessCount: 20"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalPayloadSize: 200 B"))
	assert.True(t, strings.Contains(output, "AveragePayloadSize: 10 B"))
}

func TestBlastWithLoadGenerationForMaximumDuration(t *testing.T) {
	payloadSizeBytes := int64(10)
	server, err := NewEchoServer("tcp", "localhost:10002", payloadSizeBytes)
	assert.Nil(t, err)

	server.accept(t)
	defer server.stop()

	concurrency, totalRequests := uint(1000), uint(2_00_000)

	groupOptions := workers.NewGroupOptionsWithConnections(
		concurrency,
		10,
		totalRequests,
		[]byte("HelloWorld"),
		"localhost:10002",
	)
	buffer := &bytes.Buffer{}
	blast.OutputStream = buffer
	blast.NewBlastWithoutResponseReading(groupOptions, 10*time.Millisecond)

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))

	regexp := regexp.MustCompile("TotalRequests.*")
	totalRequestsString := regexp.Find(buffer.Bytes())
	totalRequestsMade, _ := strconv.Atoi(strings.Trim(
		strings.ReplaceAll(string(totalRequestsString), "TotalRequests:", ""),
		" ",
	))

	assert.True(t, totalRequestsMade < 2_00_000)
}
