package report

import (
	"time"

	"blast/workers"
)

type Report struct {
	loadMetrics     LoadMetrics
	responseMetrics ResponseMetrics
}

// TODO: Total connections
type LoadMetrics struct {
	successCount         uint
	errorCount           uint
	errorCountByType     map[string]uint
	totalPayloadLength   int64
	averagePayloadLength float64
	earliestLoadSendTime time.Time
	latestLoadSendTime   time.Time
}

type ResponseMetrics struct {
	successCount                 uint
	errorCount                   uint
	errorCountByType             map[string]uint
	totalResponsePayloadLength   int64
	averageResponsePayloadLength float64
	earliestResponseReceivedTime time.Time
	latestResponseReceivedTime   time.Time
}

type Reporter struct {
	report                *Report
	loadGenerationChannel chan workers.LoadGenerationResponse
	responseChannel       chan SubjectServerResponse
}

func NewReporter(
	loadGenerationChannel chan workers.LoadGenerationResponse,
	responseChannel chan SubjectServerResponse,
) Reporter {
	return Reporter{
		report: &Report{
			loadMetrics: LoadMetrics{
				errorCountByType: make(map[string]uint),
			},
			responseMetrics: ResponseMetrics{
				errorCountByType: make(map[string]uint),
			},
		},
		loadGenerationChannel: loadGenerationChannel,
		responseChannel:       responseChannel,
	}
}

func (reporter *Reporter) Run() {
	reporter.collectLoadMetrics()
	reporter.collectResponseMetrics()
}

func (reporter *Reporter) collectLoadMetrics() {
	go func() {
		totalGeneratedLoad := 0
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++

			if load.Err != nil {
				reporter.report.loadMetrics.errorCount++
				reporter.report.loadMetrics.errorCountByType[load.Err.Error()]++
			} else {
				reporter.report.loadMetrics.successCount++
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

func (reporter *Reporter) collectResponseMetrics() {
	go func() {
		totalResponses := 0
		for response := range reporter.responseChannel {
			totalResponses++

			if response.Err != nil {
				reporter.report.responseMetrics.errorCount++
				reporter.report.responseMetrics.errorCountByType[response.Err.Error()]++
			} else {
				reporter.report.responseMetrics.successCount++
			}
			reporter.report.responseMetrics.totalResponsePayloadLength += response.PayloadLength
			reporter.report.responseMetrics.averageResponsePayloadLength = float64(
				reporter.report.responseMetrics.totalResponsePayloadLength,
			) / float64(totalResponses)

			if reporter.report.responseMetrics.earliestResponseReceivedTime.IsZero() ||
				response.ResponseTime.Before(
					reporter.report.responseMetrics.earliestResponseReceivedTime,
				) {
				reporter.report.responseMetrics.earliestResponseReceivedTime = response.ResponseTime
			}

			if reporter.report.responseMetrics.latestResponseReceivedTime.IsZero() ||
				response.ResponseTime.After(
					reporter.report.responseMetrics.latestResponseReceivedTime,
				) {
				reporter.report.responseMetrics.latestResponseReceivedTime = response.ResponseTime
			}
		}
	}()
}
