package blast

import (
	"io"
	"os"
	"time"

	"blast/report"
	"blast/workers"
)

var OutputStream io.Writer = os.Stdout

const MaxResponsesToRead = 10_00_000

type ResponseOptions struct {
	ResponsePayloadSizeBytes       int64
	TotalResponsesToRead           uint
	TotalSuccessfulResponsesToRead uint
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
) {
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

	blast := setUpBlast()
	blast.runWithoutResponseReading()

	<-blast.doneChannel
	blast.reporter.PrintReport(OutputStream)
}

func NewBlastWithResponseReading(
	workerGroupOptions workers.GroupOptions,
	responseOptions ResponseOptions,
	maxRunDuration time.Duration,
) {
	newResponseReader := func() (*report.ResponseReader, chan report.SubjectServerResponse) {
		responseChannel := make(chan report.SubjectServerResponse, MaxResponsesToRead)
		return report.NewResponseReader(
				responseOptions.ResponsePayloadSizeBytes,
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

	blast := setUpBlast()
	blast.runwWithResponseReading()

	<-blast.doneChannel
	blast.reporter.PrintReport(OutputStream)
}

func (blast Blast) runWithoutResponseReading() {
	loadReportedInspectionTimer := time.NewTicker(5 * time.Millisecond)
	maxRunTimer := time.NewTimer(blast.maxRunDuration)

	go func() {
		stopAll := func() {
			blast.workerGroup.Close()
			loadReportedInspectionTimer.Stop()
			maxRunTimer.Stop()
			close(blast.loadGenerationResponseChannel)
			close(blast.doneChannel)
		}

		for {
			select {
			case <-blast.workerGroup.DoneChannel():
				println("load completed")
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
			}
		}
	}()
}

func (blast Blast) runwWithResponseReading() {
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
			close(blast.doneChannel)
		}

		// TODO: ensure that totalresponsestoread >= totalrequests
		for {
			select {
			case <-blast.workerGroup.DoneChannel():
				println("load completed")
			case <-responsesCapturedInspectionTimer.C:
				if blast.responseReader.TotalResponsesRead() >= uint64(
					blast.responseOptions.TotalResponsesToRead,
				) || blast.responseReader.TotalSuccessfulResponsesRead() >= uint64(
					blast.responseOptions.TotalSuccessfulResponsesToRead,
				) {
					stopAll()
					return
				}
			case <-maxRunTimer.C:
				stopAll()
				return
			}
		}
	}()
}
