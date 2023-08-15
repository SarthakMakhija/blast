package report

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"blast/workers"
)

func TestReportWithErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel, nil)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.loadMetrics.errorCount)
	assert.Equal(t, map[string]uint{"test error": 1}, reporter.report.loadMetrics.errorCountByType)
}

func TestReportWithoutErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel, nil)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 15,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.loadMetrics.successCount)
}

func TestReportWithAndWithoutErrorInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel, nil)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 15,
	}
	loadGenerationChannel <- workers.LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.loadMetrics.successCount)
	assert.Equal(t, uint(1), reporter.report.loadMetrics.errorCount)
}

func TestReportWithPayloadLengthInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel, nil)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 10,
	}
	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, int64(20), reporter.report.loadMetrics.totalPayloadLengthBytes)
	assert.Equal(t, float64(10.0), reporter.report.loadMetrics.averagePayloadLengthBytes)
}

func TestReportWithLoadTimeInGeneratingLoad(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel, nil)
	reporter.Run()

	now := time.Now()
	laterByTenSeconds := now.Add(10 * time.Second)

	loadGenerationChannel <- workers.LoadGenerationResponse{
		LoadGenerationTime: now,
	}
	loadGenerationChannel <- workers.LoadGenerationResponse{
		LoadGenerationTime: laterByTenSeconds,
	}

	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, now, reporter.report.loadMetrics.earliestLoadSendTime)
	assert.Equal(t, laterByTenSeconds, reporter.report.loadMetrics.latestLoadSendTime)
}

func TestReportWithErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.responseMetrics.errorCount)
	assert.Equal(
		t,
		map[string]uint{"test error": 1},
		reporter.report.responseMetrics.errorCountByType,
	)
}

func TestReportWithoutErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		Response: []byte("response"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.responseMetrics.successCount)
}

func TestReportWithAndWithoutErrorInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		Response: []byte("response"),
	}
	responseChannel <- SubjectServerResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, uint(1), reporter.report.responseMetrics.successCount)
	assert.Equal(t, uint(1), reporter.report.responseMetrics.errorCount)
}

func TestReportWithResponsePayloadLengthInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewReporter(nil, responseChannel)
	reporter.Run()

	responseChannel <- SubjectServerResponse{
		PayloadLength: 10,
	}
	responseChannel <- SubjectServerResponse{
		PayloadLength: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(responseChannel)

	assert.Equal(t, int64(20), reporter.report.responseMetrics.totalResponsePayloadLengthBytes)
	assert.Equal(
		t,
		float64(10.0),
		reporter.report.responseMetrics.averageResponsePayloadLengthBytes,
	)
}

func TestReportWithLoadTimeInReceivingResponse(t *testing.T) {
	responseChannel := make(chan SubjectServerResponse, 1)
	reporter := NewReporter(nil, responseChannel)
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

	assert.Equal(t, now, reporter.report.responseMetrics.earliestResponseReceivedTime)
	assert.Equal(t, laterByTenSeconds, reporter.report.responseMetrics.latestResponseReceivedTime)
}
