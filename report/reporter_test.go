package report

import (
	"errors"
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

	assert.Equal(t, uint(1), reporter.report.Load.SuccessCount)
	assert.Equal(t, uint(1), reporter.report.Load.ErrorCount)
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
	loadGenerationChannel := make(chan LoadGenerationResponse, 1)
	reporter := NewLoadGenerationMetricsCollectingReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}
	loadGenerationChannel <- LoadGenerationResponse{
		PayloadLengthBytes: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, int64(20), reporter.report.Load.TotalPayloadLengthBytes)
	assert.Equal(t, float64(10.0), reporter.report.Load.AveragePayloadLengthBytes)
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
		Response: []byte("response"),
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
		Response: []byte("response"),
	}
	responseChannel <- SubjectServerResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.Response.SuccessCount)
	assert.Equal(t, uint(1), reporter.report.Response.ErrorCount)
}

func TestReportWithResponsePayloadLengthInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
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

	assert.Equal(t, int64(20), reporter.report.Response.TotalResponsePayloadLengthBytes)
	assert.Equal(
		t,
		float64(10.0),
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
