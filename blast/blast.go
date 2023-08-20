package blast

import (
	"fmt"
	"io"
	"os"
	"time"

	"blast/report"
	"blast/workers"
)

var OutputStream io.Writer = os.Stdout

const MaxResponsesToRead = 10_00_000

type ResponseReadingOption uint8

const (
	ReadTotalResponses      ResponseReadingOption = iota
	ReadSuccessfulResponses                       = 1
)

type ResponseOptions struct {
	ResponsePayloadSizeBytes       int64
	TotalResponsesToRead           uint
	TotalSuccessfulResponsesToRead uint
	ReadingOption                  ResponseReadingOption
	ReadDeadline                   time.Duration
}

type Blast struct {
	reporter                      *report.Reporter
	responseReader                *report.ResponseReader
	groupOptions                  workers.GroupOptions
	responseOptions               ResponseOptions
	workerGroup                   *workers.WorkerGroup
	loadGenerationResponseChannel chan report.LoadGenerationResponse
	responseChannel               chan report.SubjectServerResponse
	doneChannel                   chan struct{}
	maxRunDuration                time.Duration
}

func NewBlastWithoutResponseReading(
	workerGroupOptions workers.GroupOptions,
	maxRunDuration time.Duration,
) Blast {
	startLoad := func() (*workers.WorkerGroup, chan report.LoadGenerationResponse) {
		workerGroup := workers.NewWorkerGroup(workerGroupOptions)
		return workerGroup, workerGroup.Run()
	}

	startReporter := func(loadGenerationResponseChannel chan report.LoadGenerationResponse) *report.Reporter {
		reporter := report.
			NewLoadGenerationMetricsCollectingReporter(loadGenerationResponseChannel)

		reporter.Run()
		return reporter
	}

	setUpBlast := func() Blast {
		workerGroup, loadGenerationResponseChannel := startLoad()
		reporter := startReporter(loadGenerationResponseChannel)

		return Blast{
			reporter:                      reporter,
			groupOptions:                  workerGroupOptions,
			workerGroup:                   workerGroup,
			loadGenerationResponseChannel: loadGenerationResponseChannel,
			doneChannel:                   make(chan struct{}),
			maxRunDuration:                maxRunDuration,
		}
	}

	return setUpBlast()
}

func NewBlastWithResponseReading(
	workerGroupOptions workers.GroupOptions,
	responseOptions ResponseOptions,
	maxRunDuration time.Duration,
) Blast {
	newResponseReader := func() (*report.ResponseReader, chan report.SubjectServerResponse) {
		responseChannel := make(chan report.SubjectServerResponse, MaxResponsesToRead)
		return report.NewResponseReader(
				responseOptions.ResponsePayloadSizeBytes,
				responseOptions.ReadDeadline,
				responseChannel,
			),
			responseChannel
	}

	startLoad := func(responseReader *report.ResponseReader) (*workers.WorkerGroup, chan report.LoadGenerationResponse) {
		workerGroup := workers.NewWorkerGroupWithResponseReader(workerGroupOptions, responseReader)
		return workerGroup, workerGroup.Run()
	}

	startReporter := func(
		loadGenerationResponseChannel chan report.LoadGenerationResponse,
		responseChannel chan report.SubjectServerResponse,
	) *report.Reporter {
		reporter := report.
			NewResponseMetricsCollectingReporter(loadGenerationResponseChannel, responseChannel)

		reporter.Run()
		return reporter
	}

	setUpBlast := func() Blast {
		responseReader, responseChannel := newResponseReader()
		workerGroup, loadGenerationResponseChannel := startLoad(responseReader)
		reporter := startReporter(loadGenerationResponseChannel, responseChannel)

		return Blast{
			reporter:                      reporter,
			responseReader:                responseReader,
			responseOptions:               responseOptions,
			workerGroup:                   workerGroup,
			loadGenerationResponseChannel: loadGenerationResponseChannel,
			responseChannel:               responseChannel,
			doneChannel:                   make(chan struct{}),
			maxRunDuration:                maxRunDuration,
		}
	}

	return setUpBlast()
}

func (blast Blast) WaitForCompletion() {
	if blast.responseReader != nil {
		blast.waitForResponsesToComplete()
	} else {
		blast.waitForLoadToComplete()
	}
	<-blast.doneChannel
	blast.reporter.PrintReport(OutputStream)
}

func (blast Blast) Stop() {
	if !isClosed(blast.doneChannel) {
		close(blast.doneChannel)
	}
}

func (blast Blast) waitForLoadToComplete() {
	loadReportedInspectionTimer := time.NewTicker(5 * time.Millisecond)
	maxRunTimer := time.NewTimer(blast.maxRunDuration)

	go func() {
		stopAll := func() {
			blast.workerGroup.Close()
			loadReportedInspectionTimer.Stop()
			maxRunTimer.Stop()
			close(blast.loadGenerationResponseChannel)
			if !isClosed(blast.doneChannel) {
				close(blast.doneChannel)
			}
		}

		for {
			select {
			case <-blast.workerGroup.DoneChannel():
				fmt.Fprintln(os.Stdout, "load completed")
			case <-loadReportedInspectionTimer.C:
				if blast.reporter.TotalLoadReportedTillNow() >= uint64(
					blast.groupOptions.TotalRequests(),
				) {
					stopAll()
					return
				}
			case <-maxRunTimer.C:
				stopAll()
				return
			case <-blast.doneChannel:
				stopAll()
				return
			}
		}
	}()
}

func (blast Blast) waitForResponsesToComplete() {
	responsesCapturedInspectionTimer := time.NewTicker(5 * time.Millisecond)
	maxRunTimer := time.NewTimer(blast.maxRunDuration)

	go func() {
		stopAll := func() {
			blast.workerGroup.Close()
			blast.responseReader.Close()
			responsesCapturedInspectionTimer.Stop()
			maxRunTimer.Stop()
			close(blast.loadGenerationResponseChannel)
			close(blast.responseChannel)
			if !isClosed(blast.doneChannel) {
				close(blast.doneChannel)
			}
		}

		for {
			select {
			case <-blast.workerGroup.DoneChannel():
				fmt.Fprintln(os.Stdout, "load completed")
			case <-responsesCapturedInspectionTimer.C:
				if blast.responseOptions.ReadingOption == ReadTotalResponses {
					if blast.responseReader.TotalResponsesRead() >= uint64(
						blast.responseOptions.TotalResponsesToRead) {
						stopAll()
						return
					}
				} else if blast.responseOptions.ReadingOption == ReadSuccessfulResponses {
					if blast.responseReader.TotalSuccessfulResponsesRead() >= uint64(
						blast.responseOptions.TotalSuccessfulResponsesToRead) {
						stopAll()
						return
					}
				}
			case <-maxRunTimer.C:
				stopAll()
				return
			case <-blast.doneChannel:
				stopAll()
				return
			}
		}
	}()
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case _, ok := <-ch:
		if !ok {
			return true
		}
		return false
	default:
	}
	return false
}
