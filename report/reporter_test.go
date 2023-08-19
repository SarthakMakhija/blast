package report

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReportWithErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.Load.ErrorCount)
	assert.Equal(t, map[string]uint{"test error": 1}, reporter.report.Load.ErrorCountByType)
}

func TestReportWithoutErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 15,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.Load.SuccessCount)
}

func TestReportWithAndWithoutErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 2)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 15,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.Load.SuccessCount)
	assert.Equal(t, uint(1), reporter.report.Load.ErrorCount)
}

func TestReportWithTotalConnections(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		ConnectionId: 1,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		ConnectionId: 2,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, uint(2), reporter.report.Load.TotalConnections)
}

func TestReportWithTotalConnectionsIncludingAnErrorInAConnection(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		ConnectionId: NilConnectionId,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		ConnectionId: 2,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, uint(1), reporter.report.Load.TotalConnections)
}

func TestReportWithTotalRequests(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	close(loadGenerationChannel)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, uint(1), reporter.report.Load.TotalRequests)
}

func TestReportWithPayloadLengthInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 2)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}

	time.Sleep(4 * time.Millisecond)
	close(loadGenerationChannel)

	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, int64(20), reporter.report.Load.TotalPayloadLengthBytes)
	assert.Equal(t, int64(10), reporter.report.Load.AveragePayloadLengthBytes)
}

func TestReportWithPayloadLengthInGeneratingLoadWithAnError(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 2)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}

	time.Sleep(4 * time.Millisecond)
	close(loadGenerationChannel)

	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, int64(10), reporter.report.Load.TotalPayloadLengthBytes)
	assert.Equal(t, int64(10), reporter.report.Load.AveragePayloadLengthBytes)
}

func TestReportWithPayloadLengthInGeneratingLoadWithAllErrors(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 2)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}

	time.Sleep(4 * time.Millisecond)
	close(loadGenerationChannel)

	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, int64(0), reporter.report.Load.TotalPayloadLengthBytes)
	assert.Equal(t, int64(0), reporter.report.Load.AveragePayloadLengthBytes)
}

func TestReportWithLoadTimeInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	now := time.Now()
	laterByTenSeconds := now.Add(10 * time.Second)

	loadGenerationChannel <- LoadGenerationResponse{
		LoadGenerationTime: now,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		LoadGenerationTime: laterByTenSeconds,
	}

	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, now, reporter.report.Load.EarliestLoadSendTime)
	assert.Equal(t, laterByTenSeconds, reporter.report.Load.LatestLoadSendTime)
}

func TestReportWithErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.Response.ErrorCount)
	assert.Equal(
		t,
		map[string]uint{"test error": 1},
		reporter.report.Response.ErrorCountByType,
	)
}

func TestReportWithoutErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		ResponseTime: time.Now(),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.Response.SuccessCount)
}

func TestReportWithAndWithoutErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		ResponseTime: time.Now(),
	}
	responseChannel <- SubjectServerResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.Response.SuccessCount)
	assert.Equal(t, uint(1), reporter.report.Response.ErrorCount)
}

func TestReportWithTotalResponses(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}
	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}

	close(responseChannel)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, uint(2), reporter.report.Response.TotalResponses)
}

func TestReportWithResponsePayloadLengthInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 2)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}
	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, int64(20), reporter.report.Response.TotalResponsePayloadLengthBytes)
	assert.Equal(
		t,
		int64(10),
		reporter.report.Response.AverageResponsePayloadLengthBytes,
	)
}

func TestReportWithResponsePayloadLengthInReceivingResponseWithAnError(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 2)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}
	responseChannel <- SubjectServerResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, int64(10), reporter.report.Response.TotalResponsePayloadLengthBytes)
	assert.Equal(
		t,
		int64(10),
		reporter.report.Response.AverageResponsePayloadLengthBytes,
	)
}

func TestReportWithResponsePayloadLengthInReceivingResponseWithAllErrors(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 2)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}
	responseChannel <- SubjectServerResponse{
		Err:                errors.New("test error"),
		PayloadLengthBytes: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, int64(0), reporter.report.Response.TotalResponsePayloadLengthBytes)
	assert.Equal(
		t,
		int64(0),
		reporter.report.Response.AverageResponsePayloadLengthBytes,
	)
}

func TestReportWithLoadTimeInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(nil, responseChannel)
	reporter.Run()

	now := time.Now()
	laterByTenSeconds := now.Add(10 * time.Second)

	responseChannel <- SubjectServerResponse{
		ResponseTime: now,
	}
	responseChannel <- SubjectServerResponse{
		ResponseTime: laterByTenSeconds,
	}

	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, now, reporter.report.Response.EarliestResponseReceivedTime)
	assert.Equal(t, laterByTenSeconds, reporter.report.Response.LatestResponseReceivedTime)
}

func TestReportWithTotalLoadReported(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 15,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint64(2), reporter.TotalLoadReportedTillNow())
}

func TestPrintsTheReportWithLoadMetricsOnly(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	buffer := &bytes.Buffer{}
	reporter.PrintReport(buffer)

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests: 1"))
	assert.True(t, strings.Contains(output, "SuccessCount: 0"))
	assert.True(t, strings.Contains(output, "ErrorCount: 1"))
}

func TestPrintsTheReportWithLoadAndResponseMetricsTogether(t *testing.T) {
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewResponseMetricsCollectingReporter(loadGenerationChannel, responseChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}
	responseChannel <- SubjectServerResponse{
		PayloadLengthBytes: 10,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)
	close(responseChannel)

	buffer := &bytes.Buffer{}
	reporter.PrintReport(buffer)

	output := string(buffer.Bytes())
	assert.True(t, strings.Contains(output, "TotalRequests: 1"))
	assert.True(t, strings.Contains(output, "SuccessCount: 1"))
	assert.True(t, strings.Contains(output, "ErrorCount: 0"))
	assert.True(t, strings.Contains(output, "TotalResponses: 1"))
}
