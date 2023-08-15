package report

import (
	"time"

	"blast/workers"
)

type Report struct {
	loadMetrics LoadMetrics
}

// TODO: Total connections
type LoadMetrics struct {
	errorCount           uint
	errorCountByType     map[string]uint
	totalPayloadLength   int64
	averagePayloadLength float64
	earliestLoadSendTime time.Time
	latestLoadSendTime   time.Time
}

type Reporter struct {
	report                *Report
	loadGenerationChannel chan workers.LoadGenerationResponse
}

func NewReporter(loadGenerationChannel chan workers.LoadGenerationResponse) Reporter {
	return Reporter{
		report: &Report{
			loadMetrics: LoadMetrics{
				errorCount:       0,
				errorCountByType: make(map[string]uint),
			},
		},
		loadGenerationChannel: loadGenerationChannel,
	}
}

func (reporter *Reporter) Run() {
	go func() {
		totalGeneratedLoad := 0
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++

			if load.Err != nil {
				reporter.report.loadMetrics.errorCount++
				reporter.report.loadMetrics.errorCountByType[load.Err.Error()]++
			}
			reporter.report.loadMetrics.totalPayloadLength += load.PayloadLength
			reporter.report.loadMetrics.averagePayloadLength = float64(
				reporter.report.loadMetrics.totalPayloadLength) / float64(totalGeneratedLoad)

			if reporter.report.loadMetrics.earliestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.Before(reporter.report.loadMetrics.earliestLoadSendTime) {
				reporter.report.loadMetrics.earliestLoadSendTime = load.LoadGenerationTime
			}
			if reporter.report.loadMetrics.latestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.After(reporter.report.loadMetrics.latestLoadSendTime) {
				reporter.report.loadMetrics.latestLoadSendTime = load.LoadGenerationTime
			}
		}
	}()
}
