package blast

import (
	"io"
	"os"
	"time"

	"blast/report"
	"blast/workers"
)

var OutputStream io.Writer = os.Stdout

type ResponseOptions struct {
	responsePayloadSizeBytes       int64
	totalResponsesToRead           uint
	totalSuccessfulResponsesToRead uint
}

type Blast struct {
	reporter                      *report.Reporter
	groupOptions                  workers.GroupOptions
	workerGroup                   *workers.WorkerGroup
	loadGenerationResponseChannel chan report.LoadGenerationResponse
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
	blast.start()

	<-blast.doneChannel
	blast.reporter.PrintReport(OutputStream)
}

func (blast Blast) start() {
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
