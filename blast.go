package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

var (
	concurrency       = flag.Uint("c", 50, "")
	connections       = flag.Uint("conn", 1, "")
	numberOfRequests  = flag.Uint("n", 1000, "")
	requestsPerSecond = flag.Float64("r", 0, "")
	requestTimeout    = flag.Int("t", 20, "")
	loadDuration      = flag.Duration("z", 0, "")
	filePath          = flag.String("f", "", "")
	cpus              = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")
)

var exitFunction = usageAndExit

var usage = `Usage: blast [options...] <url>

Options:
  -n      Number of requests to run. Default is 1000.
  -c      Number of workers to run concurrently. Total number of requests cannot
          be smaller than the concurrency level. Default is 50.
  -f      Payload file.
  -r      Rate limit, in requests per second (RPS) per worker. Default is no rate limit.
  -z      Duration of application to send requests. When duration is reached,
          application stops and exits. If duration is specified, n is ignored.
          Examples: -z 10s or -z 3m.
  -t      Timeout for each request in seconds. Default is 20 seconds, use 0 for infinite.

  -conn   Number of connections to open with the target URL.
          Total number of connections cannot be greater than the concurrency level.
          Also, concurrency level modulo connections must be equal to zero.
          Default is 1.

  -cpus   Number of cpu cores to use.
          (default for current machine is %d cores)
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	assertUrl(flag.Args()[0])
	assertFilePath(*filePath)
	assertTotalConcurrentRequestsWithClientConnections(
		*loadDuration,
		*numberOfRequests,
		*concurrency,
		*connections,
	)
	assertAndSetMaxProcs(*cpus)

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		<-interruptChannel
	}()

	if *loadDuration > 0 {
		go func() {
			time.Sleep(*loadDuration)
		}()
	}
}

func assertUrl(url string) {
	if len(strings.Trim(url, " ")) == 0 {
		exitFunction("URL cannot be blank. URL is of the form host:port.")
	}
}

func assertFilePath(filePath string) {
	if len(strings.Trim(filePath, " ")) == 0 {
		exitFunction("-f cannot be blank.")
	}
}

func assertTotalConcurrentRequestsWithClientConnections(
	loadDuration time.Duration,
	totalRequests, concurrency, connections uint,
) {
	if loadDuration > 0 {
		totalRequests = math.MaxUint
		if concurrency <= 0 {
			exitFunction("-c cannot be smaller than 1.")
		}
	} else {
		if totalRequests <= 0 || concurrency <= 0 {
			exitFunction("-n and -c cannot be smaller than 1.")
		}
		if totalRequests < concurrency {
			exitFunction("-n cannot be less than -c.")
		}
	}
	if connections <= 0 {
		exitFunction("-conn cannot be smaller than 1.")
	}
	if connections > concurrency {
		exitFunction("-conn cannot be greater than -c.")
	}
	if concurrency%connections != 0 {
		exitFunction("-c modulo -conn must be equal to zero.")
	}
}

func assertAndSetMaxProcs(cpus int) {
	if cpus <= 0 {
		exitFunction("-cpus cannot be smaller than 1.")
	}
	runtime.GOMAXPROCS(cpus)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
