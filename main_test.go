package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func exitWithPanic(msg string) {
	panic(msg)
}

func TestRunBlastWithoutUrl(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertUrl("")
	})
}

func TestRunBlastWithEmptyUrl(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertUrl(" ")
	})
}

func TestRunBlastWithoutFileAndProcessPath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertFileAndProcessPath("", "")
	})
}

func TestRunBlastWithEmptyFileAndProcessPath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertFileAndProcessPath(" ", " ")
	})
}

func TestRunBlastWithNonEmptyFilePathAndNonEmptyProcessPath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertFileAndProcessPath("./filePayload", "./exe")
	})
}

func TestRunBlastWithEmptyFileAndNonEmptyProcessPath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertFileAndProcessPath(" ", "./exe")
	})
}

func TestRunBlastWithNonEmptyFileAndEmptyProcessPath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertFileAndProcessPath("./payload", "")
	})
}

func TestBlastWithConnectTimeoutEqualToZero(t *testing.T) {
	assert.Panics(t, func() {
		assertConnectTimeout(time.Duration(0))
	})
}

func TestBlastWithConnectTimeout(t *testing.T) {
	assert.NotPanics(t, func() {
		assertConnectTimeout(time.Duration(1))
	})
}

func TestBlastWithRequestsPerSecondLessThanZero(t *testing.T) {
	assert.Panics(t, func() {
		assertRequestsPerSecond(-1)
	})
}

func TestBlastWithLoadDurationZero(t *testing.T) {
	tests := []struct {
		loadDuration string
	}{
		{loadDuration: "0s"},
		{loadDuration: "0h"},
		{loadDuration: "0ms"},
	}

	for _, test := range tests {
		duration, _ := time.ParseDuration(test.loadDuration)
		assert.Panics(t, func() {
			assertLoadDuration(duration)
		})
	}
}

func TestBlastWithRequestsPerSecond(t *testing.T) {
	assert.NotPanics(t, func() {
		assertRequestsPerSecond(1)
	})
}

func TestTotalRequestsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0, 1, 1,
		)
	})
}

func TestConcurrencyMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 0, 1,
		)
	})
}

func TestConnectionsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 1, 0,
		)
	})
}

func TestTotalRequestsMustBeGreaterThanOrEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 2, 1,
		)
	})
}

func TestTotalRequestsIsEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			2, 2, 1,
		)
	})
}

func TestTotalRequestsIsGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			4, 2, 1,
		)
	})
}

func TestConnectionsMustNotBeGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			10, 5, 10,
		)
	})
}

func TestConcurrencyMustBeAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			10, 5, 8,
		)
	})
}

func TestConcurrencyIsAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			100, 10, 5,
		)
	})
}

func TestConcurrencyMustBeGreaterThanZeroEvenIfLoadDurationIsGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 0, 1,
		)
	})
}

func TestMaxProcsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertAndSetMaxProcs(0)
	})
}

func TestMaxProcsIsGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertAndSetMaxProcs(1)
	})
}

func TestBlastWithResponseSizeLessThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertResponseReading(true, -1, 10, 10)
	})
}

func TestBlastWithBothTotalResponsesAndSuccessfulResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertResponseReading(true, 100, 10, 10)
	})
}

func TestBlastWithOnlyTotalResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 10, 0)
	})
}

func TestBlastWithOnlySuccessfulResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 0, 10)
	})
}

func TestBlastWithoutResponseReading(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 10, 10)
	})
}

func TestBlastWithNonExistingFile(t *testing.T) {
	assert.Panics(t, func() {
		getFilePayload("./non-existing")
	})
}

func TestBlastWithAnExistingFile(t *testing.T) {
	file, err := os.Create("testFile")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(file.Name())
	}()

	assert.NotPanics(t, func() {
		getFilePayload("./testFile")
	})
}
