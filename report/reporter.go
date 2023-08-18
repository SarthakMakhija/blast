package report

import (
	"io"
	"sync/atomic"
	"time"
)

type Report struct {
	Load     LoadMetrics
	Response ResponseMetrics
}

// TODO: Total connections
type LoadMetrics struct {
	TotalRequests             uint
	SuccessCount              uint
	ErrorCount                uint
	ErrorCountByType          map[string]uint
	TotalPayloadLengthBytes   int64
	AveragePayloadLengthBytes int64
	EarliestLoadSendTime      time.Time
	LatestLoadSendTime        time.Time
	TotalTime                 time.Duration
}

type ResponseMetrics struct {
	TotalResponses                    uint
	SuccessCount                      uint
	ErrorCount                        uint
	ErrorCountByType                  map[string]uint
	TotalResponsePayloadLengthBytes   int64
	AverageResponsePayloadLengthBytes int64
	EarliestResponseReceivedTime      time.Time
	LatestResponseReceivedTime        time.Time
	IsAvailableForReporting           bool
	TotalTime                         time.Duration
}

type Reporter struct {
	report                     *Report
	totalLoadReportedTillNow   atomic.Uint64
	loadGenerationChannel      chan LoadGenerationResponse
	responseChannel            chan SubjectServerResponse
	loadMetricsDoneChannel     chan struct{}
	responseMetricsDoneChannel chan struct{}
}

func NewLoadGenerationMetricsCollectingReporter(
	loadGenerationChannel chan LoadGenerationResponse,
) *Reporter {
	return &Reporter{
		report: &Report{
			Load: LoadMetrics{
				ErrorCountByType: make(map[string]uint),
			},
			Response: ResponseMetrics{
				IsAvailableForReporting: false,
			},
		},
		loadGenerationChannel:      loadGenerationChannel,
		responseChannel:            nil,
		loadMetricsDoneChannel:     make(chan struct{}),
		responseMetricsDoneChannel: nil,
	}
}

func NewResponseMetricsCollectingReporter(
	loadGenerationChannel chan LoadGenerationResponse,
	responseChannel chan SubjectServerResponse,
) *Reporter {
	return &Reporter{
		report: &Report{
			Load: LoadMetrics{
				ErrorCountByType: make(map[string]uint),
			},
			Response: ResponseMetrics{
				IsAvailableForReporting: true,
				ErrorCountByType:        make(map[string]uint),
			},
		},
		loadGenerationChannel:      loadGenerationChannel,
		responseChannel:            responseChannel,
		loadMetricsDoneChannel:     make(chan struct{}),
		responseMetricsDoneChannel: make(chan struct{}),
	}
}

func (reporter *Reporter) Run() {
	reporter.collectLoadMetrics()
	if reporter.responseChannel != nil {
		reporter.collectResponseMetrics()
	}
}

func (reporter *Reporter) PrintReport(writer io.Writer) {
	println("printing report...")
	<-reporter.loadMetricsDoneChannel
	if reporter.responseMetricsDoneChannel != nil {
		<-reporter.responseMetricsDoneChannel
	}
	print(writer, reporter.report)
}

func (reporter *Reporter) TotalLoadReportedTillNow() uint64 {
	return reporter.totalLoadReportedTillNow.Load()
}

func (reporter *Reporter) collectLoadMetrics() {
	go func() {
		totalGeneratedLoad := uint(0)
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++
			reporter.totalLoadReportedTillNow.Add(1)

			if load.Err != nil {
				reporter.report.Load.ErrorCount++
				reporter.report.Load.ErrorCountByType[load.Err.Error()]++
			} else {
				reporter.report.Load.SuccessCount++
			}
			reporter.report.Load.TotalPayloadLengthBytes += load.PayloadLengthBytes
			reporter.report.Load.AveragePayloadLengthBytes = reporter.report.Load.TotalPayloadLengthBytes / int64(
				totalGeneratedLoad,
			)

			if reporter.report.Load.EarliestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.Before(reporter.report.Load.EarliestLoadSendTime) {
				reporter.report.Load.EarliestLoadSendTime = load.LoadGenerationTime
			}

			if reporter.report.Load.LatestLoadSendTime.IsZero() ||
				load.LoadGenerationTime.After(reporter.report.Load.LatestLoadSendTime) {
				reporter.report.Load.LatestLoadSendTime = load.LoadGenerationTime
			}
		}
		startTime := reporter.report.Load.EarliestLoadSendTime
		timeToCompleteLoad := reporter.report.Load.LatestLoadSendTime.Sub(startTime)

		reporter.report.Load.TotalTime = timeToCompleteLoad
		reporter.report.Load.TotalRequests = totalGeneratedLoad
		close(reporter.loadMetricsDoneChannel)
	}()
}

func (reporter *Reporter) collectResponseMetrics() {
	go func() {
		totalResponses := 0
		for response := range reporter.responseChannel {
			totalResponses++

			if response.Err != nil {
				reporter.report.Response.ErrorCount++
				reporter.report.Response.ErrorCountByType[response.Err.Error()]++
			} else {
				reporter.report.Response.SuccessCount++
			}
			reporter.report.Response.TotalResponsePayloadLengthBytes += response.PayloadLengthBytes
			reporter.report.Response.AverageResponsePayloadLengthBytes = reporter.report.Response.TotalResponsePayloadLengthBytes /
				int64(
					totalResponses,
				)

			if reporter.report.Response.EarliestResponseReceivedTime.IsZero() ||
				response.ResponseTime.Before(
					reporter.report.Response.EarliestResponseReceivedTime,
				) {
				reporter.report.Response.EarliestResponseReceivedTime = response.ResponseTime
			}

			if reporter.report.Response.LatestResponseReceivedTime.IsZero() ||
				response.ResponseTime.After(
					reporter.report.Response.LatestResponseReceivedTime,
				) {
				reporter.report.Response.LatestResponseReceivedTime = response.ResponseTime
			}
		}
		reporter.report.Response.TotalResponses = uint(totalResponses)

		timeToCompleteResponses := reporter.report.Response.LatestResponseReceivedTime.
			Sub(reporter.report.Response.EarliestResponseReceivedTime)
		reporter.report.Response.TotalTime = timeToCompleteResponses
		close(reporter.responseMetricsDoneChannel)
	}()
}
