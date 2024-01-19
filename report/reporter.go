package report

import (
	"io"
	"sync/atomic"
	"time"
)

// Report reports represents the report that is displayed to the user after the load is completed.
// Report contains LoadMetrics and ResponseMetrics.
// LoadMetrics defines fields that are relevant to the generated load, whereas
// ResponseMetrics defines the fields that are relevant to the response read by blast.
// ResponseMetrics is only captured if NewResponseMetricsCollectingReporter method is called.
type Report struct {
	Load     LoadMetrics
	Response ResponseMetrics
}

type LoadMetrics struct {
	TotalRequests                  uint
	SuccessCount                   uint
	ErrorCount                     uint
	ErrorCountByType               map[string]uint
	TotalConnections               uint
	TotalPayloadLengthBytes        int64
	AveragePayloadLengthBytes      int64
	EarliestSuccessfulLoadSendTime time.Time
	LatestSuccessfulLoadSendTime   time.Time
	TotalTime                      time.Duration
	uniqueConnectionIds            map[int]bool
}

type ResponseMetrics struct {
	TotalResponses                         uint
	SuccessCount                           uint
	ErrorCount                             uint
	ErrorCountByType                       map[string]uint
	TotalResponsePayloadLengthBytes        int64
	AverageResponsePayloadLengthBytes      int64
	EarliestSuccessfulResponseReceivedTime time.Time
	LatestSuccessfulResponseReceivedTime   time.Time
	IsAvailableForReporting                bool
	TotalTime                              time.Duration
}

// Reporter generates the report.
// It is implemented as two goroutines, one that listens to the loadGenerationChannel and other
// that listens to the responseChannel.
// One goroutine populates the LoadMetrics, and the other goroutine populates the ResponseMetrics.
type Reporter struct {
	report                     *Report
	totalLoadReportedTillNow   atomic.Uint64
	loadGenerationChannel      chan LoadGenerationResponse
	responseChannel            chan SubjectServerResponse
	loadMetricsDoneChannel     chan struct{}
	responseMetricsDoneChannel chan struct{}
}

// NewLoadGenerationMetricsCollectingReporter creates a new Reporter that only populates
// the LoadMetrics.
func NewLoadGenerationMetricsCollectingReporter(
	loadGenerationChannel chan LoadGenerationResponse,
) *Reporter {
	return &Reporter{
		report: &Report{
			Load: LoadMetrics{
				ErrorCountByType:    make(map[string]uint),
				uniqueConnectionIds: make(map[int]bool),
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

// NewResponseMetricsCollectingReporter creates a new Reporter that populates both the
// LoadMetrics and ResponseMetrics.
func NewResponseMetricsCollectingReporter(
	loadGenerationChannel chan LoadGenerationResponse,
	responseChannel chan SubjectServerResponse,
) *Reporter {
	return &Reporter{
		report: &Report{
			Load: LoadMetrics{
				ErrorCountByType:    make(map[string]uint),
				uniqueConnectionIds: make(map[int]bool),
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

// Run runs the Reporter goroutines.
func (reporter *Reporter) Run() {
	reporter.collectLoadMetrics()
	if reporter.responseChannel != nil {
		reporter.collectResponseMetrics()
	}
}

// PrintReport prints the report on the provided io.Writer.
// Before the report is ready to be printed, PrintReport waits for
// the goroutines to finish.
// This method can only be called after the loadGenerationChannel and responseChannel
// are closed.
func (reporter *Reporter) PrintReport(writer io.Writer) {
	<-reporter.loadMetricsDoneChannel
	if reporter.responseMetricsDoneChannel != nil {
		<-reporter.responseMetricsDoneChannel
	}
	_ = write(writer, reporter.report)
}

// TotalLoadReportedTillNow returns the total load that has reporter so far.
func (reporter *Reporter) TotalLoadReportedTillNow() uint64 {
	return reporter.totalLoadReportedTillNow.Load()
}

// collectLoadMetrics runs a goroutine that reports the LoadMetrics.
func (reporter *Reporter) collectLoadMetrics() {
	go func() {
		totalGeneratedLoad := uint(0)
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++
			reporter.totalLoadReportedTillNow.Add(1)

			if load.ConnectionId != NilConnectionId {
				reporter.report.Load.uniqueConnectionIds[load.ConnectionId] = true
			}

			if load.Err != nil {
				reporter.report.Load.ErrorCount++
				reporter.report.Load.ErrorCountByType[load.Err.Error()]++
			} else {
				reporter.report.Load.SuccessCount++
				reporter.report.Load.TotalPayloadLengthBytes += load.PayloadLengthBytes

				if reporter.report.Load.EarliestSuccessfulLoadSendTime.IsZero() ||
					load.LoadGenerationTime.Before(reporter.report.Load.EarliestSuccessfulLoadSendTime) {
					reporter.report.Load.EarliestSuccessfulLoadSendTime = load.LoadGenerationTime
				}

				if reporter.report.Load.LatestSuccessfulLoadSendTime.IsZero() ||
					load.LoadGenerationTime.After(reporter.report.Load.LatestSuccessfulLoadSendTime) {
					reporter.report.Load.LatestSuccessfulLoadSendTime = load.LoadGenerationTime
				}
			}
		}
		startTime := reporter.report.Load.EarliestSuccessfulLoadSendTime
		timeToCompleteLoad := reporter.report.Load.LatestSuccessfulLoadSendTime.Sub(startTime)

		if reporter.report.Load.SuccessCount != 0 {
			reporter.report.Load.AveragePayloadLengthBytes = reporter.report.Load.TotalPayloadLengthBytes / int64(
				reporter.report.Load.SuccessCount,
			)
		} else {
			reporter.report.Load.AveragePayloadLengthBytes = 0
		}
		reporter.report.Load.TotalTime = timeToCompleteLoad
		reporter.report.Load.TotalRequests = totalGeneratedLoad
		reporter.report.Load.TotalConnections = uint(len(reporter.report.Load.uniqueConnectionIds))

		close(reporter.loadMetricsDoneChannel)
	}()
}

// collectResponseMetrics runs a goroutine that reports the ResponseMetrics.
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
				reporter.report.Response.TotalResponsePayloadLengthBytes += response.PayloadLengthBytes

				if reporter.report.Response.EarliestSuccessfulResponseReceivedTime.IsZero() ||
					response.ResponseTime.Before(
						reporter.report.Response.EarliestSuccessfulResponseReceivedTime,
					) {
					reporter.report.Response.EarliestSuccessfulResponseReceivedTime = response.ResponseTime
				}

				if reporter.report.Response.LatestSuccessfulResponseReceivedTime.IsZero() ||
					response.ResponseTime.After(
						reporter.report.Response.LatestSuccessfulResponseReceivedTime,
					) {
					reporter.report.Response.LatestSuccessfulResponseReceivedTime = response.ResponseTime
				}
			}
		}
		reporter.report.Response.TotalResponses = uint(totalResponses)
		if reporter.report.Response.SuccessCount != 0 {
			reporter.report.Response.AverageResponsePayloadLengthBytes = reporter.report.Response.TotalResponsePayloadLengthBytes /
				int64(
					reporter.report.Response.SuccessCount,
				)
		} else {
			reporter.report.Response.AverageResponsePayloadLengthBytes = 0
		}

		timeToCompleteResponses := reporter.report.Response.LatestSuccessfulResponseReceivedTime.
			Sub(reporter.report.Response.EarliestSuccessfulResponseReceivedTime)
		reporter.report.Response.TotalTime = timeToCompleteResponses

		close(reporter.responseMetricsDoneChannel)
	}()
}
