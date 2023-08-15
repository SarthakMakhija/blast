package report

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"blast/workers"
)

func TestReportWithError(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		Err: errors.New("test error"),
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.loadMetrics.errorCount)
	assert.Equal(t, map[string]uint{"test error": 1}, reporter.report.loadMetrics.errorCountByType)
}

func TestReportWithoutError(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 15,
	}
	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, uint(1), reporter.report.loadMetrics.successCount)
}

func TestReportWithAndWithoutError(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel)
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

func TestReportWithPayloadLength(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel)
	reporter.Run()

	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 10,
	}
	loadGenerationChannel <- workers.LoadGenerationResponse{
		PayloadLength: 10,
	}

	time.Sleep(2 * time.Millisecond)
	close(loadGenerationChannel)

	assert.Equal(t, int64(20), reporter.report.loadMetrics.totalPayloadLength)
	assert.Equal(t, float64(10.0), reporter.report.loadMetrics.averagePayloadLength)
}

func TestReportWithLoadTime(t *testing.T) {
	loadGenerationChannel := make(chan workers.LoadGenerationResponse, 1)
	reporter := NewReporter(loadGenerationChannel)
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
