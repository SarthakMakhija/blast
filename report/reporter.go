package report

import (
	"io"
	"time"
)

type Report struct {
	Load     LoadMetrics
	Response ResponseMetrics
}

// TODO: Total connections
type LoadMetrics struct {
	TotalRequests             uint
	RequestsPerSecond         float64
	SuccessCount              uint
	ErrorCount                uint
	ErrorCountByType          map[string]uint
	TotalPayloadLengthBytes   int64
	AveragePayloadLengthBytes float64
	EarliestLoadSendTime      time.Time
	LatestLoadSendTime        time.Time
	TotalTime                 time.Duration
}

// TODO: connection time, total responses, time to get the responses
type ResponseMetrics struct {
	SuccessCount                      uint
	ErrorCount                        uint
	ErrorCountByType                  map[string]uint
	TotalResponsePayloadLengthBytes   int64
	AverageResponsePayloadLengthBytes float64
	EarliestResponseReceivedTime      time.Time
	LatestResponseReceivedTime        time.Time
	IsAvailableForReporting           bool
}

type Reporter struct {
	report                *Report
	loadGenerationChannel chan LoadGenerationResponse
	responseChannel       chan SubjectServerResponse
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
		loadGenerationChannel: loadGenerationChannel,
		responseChannel:       nil,
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
		loadGenerationChannel: loadGenerationChannel,
		responseChannel:       responseChannel,
	}
}

func (reporter *Reporter) Run() {
	reporter.collectLoadMetrics()
	if reporter.responseChannel != nil {
		reporter.collectResponseMetrics()
	}
}

func (reporter *Reporter) PrintReport(writer io.Writer) {
	print(writer, reporter.report)
}

func (reporter *Reporter) collectLoadMetrics() {
	go func() {
		totalGeneratedLoad := uint(0)
		for load := range reporter.loadGenerationChannel {
			totalGeneratedLoad++

			if load.Err != nil {
				reporter.report.Load.ErrorCount++
				reporter.report.Load.ErrorCountByType[load.Err.Error()]++
			} else {
				reporter.report.Load.SuccessCount++
			}
			reporter.report.Load.TotalPayloadLengthBytes += load.PayloadLengthBytes
			reporter.report.Load.AveragePayloadLengthBytes = float64(
				reporter.report.Load.TotalPayloadLengthBytes,
			) / float64(
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
		timeToCompleteLoad := time.Now().Sub(startTime)

		reporter.report.Load.TotalTime = timeToCompleteLoad
		reporter.report.Load.TotalRequests = totalGeneratedLoad
		reporter.report.Load.RequestsPerSecond = float64(
			totalGeneratedLoad,
		) / float64(
			timeToCompleteLoad.Seconds(),
		)
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
			reporter.report.Response.AverageResponsePayloadLengthBytes = float64(
				reporter.report.Response.TotalResponsePayloadLengthBytes,
			) / float64(totalResponses)

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
	}()
}
