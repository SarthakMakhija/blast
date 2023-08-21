package blast

import (
	"fmt"
	"io"
	"os"
	"time"

	"blast/report"
	"blast/workers"
)

// OutputStream defines a io.Writer to write the report to.
// Currently, os.Stdout is the only supported io.Writer.
// The entire system writes the error messages to os.Stderr.
var OutputStream io.Writer = os.Stdout

// MaxResponsesToRead is the size of the responseChannel on which the responses read from the server are sent.
const MaxResponsesToRead = 10_00_000

// ResponseReadingOption defines the type of responses to read before the load on the target server stops.
// 1) Total responses
// 2) Successful responses
type ResponseReadingOption uint8

const (
	ReadTotalResponses      ResponseReadingOption = iota
	ReadSuccessfulResponses                       = 1
)

// ResponseOptions defines the options for reading responses from the target server.
type ResponseOptions struct {
	ResponsePayloadSizeBytes       int64
	TotalResponsesToRead           uint
	TotalSuccessfulResponsesToRead uint
	ReadingOption                  ResponseReadingOption
	ReadDeadline                   time.Duration
}

// Blast runs the workers for sending the load, starting the reporters and waiting for the process to complete.
// It orchestrates between workers.WorkerGroup, report.Reporter and report.ResponseReader.
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

// NewBlastWithoutResponseReading returns a new instance of Blast that does not read responses from the target server.
func NewBlastWithoutResponseReading(
	workerGroupOptions workers.GroupOptions,
	maxRunDuration time.Duration,
) Blast {
	// startLoad starts the workers for sending load on the target server.
	startLoad := func() (*workers.WorkerGroup, chan report.LoadGenerationResponse) {
		workerGroup := workers.NewWorkerGroup(workerGroupOptions)
		return workerGroup, workerGroup.Run()
	}

	// startReporter starts the reporter.
	startReporter := func(loadGenerationResponseChannel chan report.LoadGenerationResponse) *report.Reporter {
		reporter := report.
			NewLoadGenerationMetricsCollectingReporter(loadGenerationResponseChannel)

		reporter.Run()
		return reporter
	}

	// setUpBlast creates a new instance of Blast.
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

// NewBlastWithResponseReading creates a new instance of Blast that reads responses from the target server.
func NewBlastWithResponseReading(
	workerGroupOptions workers.GroupOptions,
	responseOptions ResponseOptions,
	maxRunDuration time.Duration,
) Blast {
	// newResponseReader creates a new instance of ResponseReader that reads responses from the target server.
	newResponseReader := func() (*report.ResponseReader, chan report.SubjectServerResponse) {
		responseChannel := make(chan report.SubjectServerResponse, MaxResponsesToRead)
		return report.NewResponseReader(
				responseOptions.ResponsePayloadSizeBytes,
				responseOptions.ReadDeadline,
				responseChannel,
			),
			responseChannel
	}

	// startLoad starts the workers for sending load on the target server.
	startLoad := func(responseReader *report.ResponseReader) (*workers.WorkerGroup, chan report.LoadGenerationResponse) {
		workerGroup := workers.NewWorkerGroupWithResponseReader(workerGroupOptions, responseReader)
		return workerGroup, workerGroup.Run()
	}

	// startReporter starts the reporter.
	startReporter := func(
		loadGenerationResponseChannel chan report.LoadGenerationResponse,
		responseChannel chan report.SubjectServerResponse,
	) *report.Reporter {
		reporter := report.
			NewResponseMetricsCollectingReporter(loadGenerationResponseChannel, responseChannel)

		reporter.Run()
		return reporter
	}

	// setUpBlast creates a new instance of Blast.
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

// WaitForCompletion waits for the load to complete.
// Case1:
// Consider that Blast is configured to run without response reading. In this case, WaitForCompletion will finish:
// If either the load is complete => TotalRequests have been sent to the target server or,
// Blast has run for the specified maximum duration or,
// Blast is made to stop.
// Case2:
// Consider that Blast is configured to run with response reading. In this case, WaitForCompletion will finish:
// If either the load is complete => TotalRequests have been sent to the target server or,
// The total responses or the total successful responses have been read from the target server or,
// Blast has run for the specified maximum duration.
// Blast is made to stop.
func (blast Blast) WaitForCompletion() {
	if blast.responseReader != nil {
		blast.waitForResponsesToComplete()
	} else {
		blast.waitForLoadToComplete()
	}
	<-blast.doneChannel
	blast.reporter.PrintReport(OutputStream)
}

// Stop stops the blast, usually called when an interrupt is received in ../main.go.
func (blast Blast) Stop() {
	if !isClosed(blast.doneChannel) {
		close(blast.doneChannel)
	}
}

// waitForLoadToComplete finishes if either of the conditions are true:
// If the load is complete => TotalRequests have been sent to the target server or,
// Blast has run for the specified maximum duration or,
// Blast is made to stop.
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
				fmt.Fprintln(os.Stdout, "[Load completed]")
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

// waitForResponsesToComplete finishes if either of the conditions are true:
// If either the load is complete => TotalRequests have been sent to the target server or,
// The total responses or the total successful responses have been read from the target server or,
// Blast has run for the specified maximum duration.
// Blast is made to stop.
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
				fmt.Fprintln(os.Stdout, "[Load completed]")
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

// isClosed returns true if the channel is closed, false otherwise.
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
