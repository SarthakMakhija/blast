package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/dimiro1/banner"

	"blast/cmd"
	payloadprovider "blast/payload_provider"
	"blast/workers"
)

var (
	numberOfRequests        = flag.Uint("n", 1000, "")
	concurrency             = flag.Uint("c", 50, "")
	connections             = flag.Uint("conn", 1, "")
	filePath                = flag.String("f", "", "")
	requestsPerSecond       = flag.Float64("rps", 0, "")
	loadDuration            = flag.Duration("z", 20*time.Second, "")
	connectTimeout          = flag.Duration("t", 3*time.Second, "")
	readResponses           = flag.Bool("Rr", false, "")
	responsePayloadSize     = flag.Int64("Rrs", -1, "")
	readResponseDeadline    = flag.Duration("Rrd", 0*time.Second, "")
	readTotalResponses      = flag.Uint("Rtr", 0, "")
	readSuccessfulResponses = flag.Uint("Rsr", 0, "")
	cpus                    = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")
)

const (
	version      = "0.0.2"
	versionLabel = " Version: %s\n\n"
)

var exitFunction = usageAndExit

var usage = `blast is a load generator for TCP servers which maintain persistent connections.

Usage: blast [options...] <url>

Options:
  -n      Number of requests to run. Default is 1000.
  -c      Number of workers to run concurrently. Total number of requests cannot
          be smaller than the concurrency level. Default is 50.
  -f      File path containing the load payload.
  -rps    Rate limit in requests per second (RPS) per worker. Default is no rate limit.
  -z      Duration of blast to send requests. When duration is reached,
          application stops and exits. Default is 20 seconds.
          Example usage: -z 10s or -z 3m.
  -t      Timeout for establishing connection with the target server. Default is 3 seconds.
          Also called as DialTimeout.
  -Rr     Read responses from the target server. Default is false.
  -Rrs    Read response size is the size of the responses in bytes returned by the target server. 
  -Rrd    Read response deadline defines the deadline for the read calls on connection.
          Default is no deadline which means the read calls do not timeout.
          This flag is applied only if "Read responses" (-Rr) is true.
  -Rtr    Read total responses is the total responses to read from the target server. 
          The load generation will stop if either the duration (-z) has exceeded or the total 
          responses have been read. This flag is applied only if "Read responses" (-Rr)
          is true.
  -Rsr    Read successful responses  is the successful responses to read from the target server. 
          The load generation will stop if either the duration (-z) has exceeded or 
          the total successful responses have been read. Either of "-Rtr"
          or "-Rsr" must be specified, if -Rr is set. This flag is applied only if 
          "Read responses" (-Rr) is true.

  -conn   Number of connections to open with the target URL.
          Total number of connections cannot be greater than the concurrency level.
          Also, concurrency level modulo connections must be equal to zero.
          Default is 1.

  -cpus   Number of cpu cores to use.
          (default for current machine is %d cores)
`

// main is the CLI for the blast application.
func main() {
	logo := `{{ .Title "blast" "" 0}}`
	banner.InitString(os.Stdout, true, false, logo)
	fmt.Fprintf(os.Stdout, versionLabel, version)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	url := flag.Args()[0]
	assertUrl(url)
	assertFilePath(*filePath)
	assertConnectTimeout(*connectTimeout)
	assertRequestsPerSecond(*requestsPerSecond)
	assertLoadDuration(*loadDuration)
	assertTotalConcurrentRequestsWithClientConnections(
		*numberOfRequests,
		*concurrency,
		*connections,
	)
	assertResponseReading(
		*readResponses,
		*responsePayloadSize,
		*readTotalResponses,
		*readSuccessfulResponses,
	)
	assertAndSetMaxProcs(*cpus)

	blastInstance := setUpBlast(url)

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	go func() {
		<-interruptChannel
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()
}

// assertUrl asserts that the URL is not empty.
func assertUrl(url string) {
	if len(strings.Trim(url, " ")) == 0 {
		exitFunction("URL cannot be blank. URL is of the form host:port.")
	}
}

// assertFilePath asserts that the filePath is not empty.
func assertFilePath(filePath string) {
	if len(strings.Trim(filePath, " ")) == 0 {
		exitFunction("-f cannot be blank.")
	}
}

// assertConnectTimeout asserts that the connectTimeout is greater than zero.
func assertConnectTimeout(timeout time.Duration) {
	if timeout <= time.Duration(0) {
		exitFunction("-t cannot be smaller than or equal to zero.")
	}
}

// assertRequestsPerSecond asserts that the requestsPerSecond is greater than or equal to zero.
func assertRequestsPerSecond(requestsPerSecond float64) {
	if requestsPerSecond < 0 {
		exitFunction("-rps cannot be smaller than zero.")
	}
}

// assertLoadDuration asserts that the loadDuration is greater than zero.
func assertLoadDuration(duration time.Duration) {
	if duration <= time.Duration(0) {
		exitFunction("-z cannot be smaller than or equal to zero.")
	}
}

// assertTotalConcurrentRequestsWithClientConnections asserts the relationship between concurrency, totalRequests and
// client connections.
func assertTotalConcurrentRequestsWithClientConnections(
	totalRequests, concurrency, connections uint,
) {
	if connections <= 0 {
		exitFunction("-conn cannot be smaller than 1.")
	}
	if totalRequests <= 0 || concurrency <= 0 {
		exitFunction("-n and -c cannot be smaller than 1.")
	}
	if totalRequests < concurrency {
		exitFunction("-n cannot be smaller than -c.")
	}
	if connections > concurrency {
		exitFunction("-conn cannot be greater than -c.")
	}
	if concurrency%connections != 0 {
		exitFunction("-c modulo -conn must be equal to zero.")
	}
}

// assertAndSetMaxProcs asserts the maximum number of cpus and sets the value in GOMAXPROCS.
func assertAndSetMaxProcs(cpus int) {
	if cpus <= 0 {
		exitFunction("-cpus cannot be smaller than 1.")
	}
	runtime.GOMAXPROCS(cpus)
}

// assertResponseReading asserts the options related to reading responses.
func assertResponseReading(
	readResponses bool,
	responsePayloadSize int64,
	readTotalResponses, readSuccessfulResponses uint,
) {
	if readResponses {
		if responsePayloadSize < 0 {
			exitFunction("-Rrs cannot be smaller than 0.")
		}
		if readTotalResponses > 0 && readSuccessfulResponses > 0 {
			exitFunction("both -Rtr and -Rsr cannot be specified.")
		}
		if readTotalResponses == 0 && readSuccessfulResponses == 0 {
			exitFunction("either of -Rtr or -Rsr must be specified.")
		}
	}
}

// setUpBlast creates a new instance of blast.Blast.
func setUpBlast(url string) blast.Blast {
	groupOptions := workers.NewGroupOptionsFullyLoaded(
		*concurrency,
		*connections,
		*numberOfRequests,
		getFilePayload(*filePath),
		url,
		*requestsPerSecond,
		*connectTimeout,
	)

	var instance blast.Blast
	if *readResponses {
		readingOption := blast.ReadTotalResponses
		if *readSuccessfulResponses > 0 {
			readingOption = blast.ReadSuccessfulResponses
		}
		responseOptions := blast.ResponseOptions{
			ResponsePayloadSizeBytes:       *responsePayloadSize,
			TotalResponsesToRead:           *readTotalResponses,
			TotalSuccessfulResponsesToRead: *readSuccessfulResponses,
			ReadingOption:                  readingOption,
			ReadDeadline:                   *readResponseDeadline,
		}
		instance = blast.NewBlastWithResponseReading(groupOptions, responseOptions, *loadDuration)
	} else {
		instance = blast.NewBlastWithoutResponseReading(groupOptions, *loadDuration)
	}
	return instance
}

// getFilePayload returns the file content.
func getFilePayload(filePath string) []byte {
	provider, err := payloadprovider.NewFilePayloadProvider(filePath)
	if err != nil {
		exitFunction(fmt.Sprintf("file path: %v does not exist.", filePath))
	}
	return provider.Get()
}

// usageAndExit defines the usage of blast application and exits the application.
func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
