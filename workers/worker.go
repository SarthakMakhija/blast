package workers

import (
	"io"
	"sync"
	"time"

	"blast/report"
)

type Worker struct {
	connection io.WriteCloser
	options    WorkerOptions
}

func (worker Worker) run(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		worker.sendRequests()
	}()
}

func (worker Worker) sendRequests() {
	var throttle <-chan time.Time
	if worker.options.requestsPerSecond > 0 {
		throttle = time.Tick(
			time.Duration(1e6/(worker.options.requestsPerSecond)) * time.Microsecond,
		)
	}

	for request := 1; request <= int(worker.options.totalRequests); request++ {
		select {
		case <-worker.options.stopChannel:
			return
		default:
			if worker.options.requestsPerSecond > 0 {
				<-throttle
			}

			worker.sendRequest()
		}
	}
}

func (worker Worker) sendRequest() {
	_, err := worker.connection.Write(worker.options.payload)
	worker.options.loadGenerationResponse <- report.LoadGenerationResponse{
		Err:                err,
		PayloadLengthBytes: int64(len(worker.options.payload)),
		LoadGenerationTime: time.Now(),
	}
}
