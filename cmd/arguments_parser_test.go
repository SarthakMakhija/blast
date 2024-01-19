package blast

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func exitWithPanic(msg string) {
	panic(msg)
}

func TestParseCommandLineArgumentsWithoutUrl(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertUrl("")
	})
}

func TestParseCommandLineArgumentsWithEmptyUrl(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertUrl(" ")
	})
}

func TestParseCommandLineArgumentsWithoutPayloadFilePath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertPayloadFilePath("")
	})
}

func TestParseCommandLineArgumentsWithEmptyPayloadFilePath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertPayloadFilePath(" ")
	})
}

func TestParseCommandLineArgumentsWithNonEmptyPayloadFilePath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertPayloadFilePath("./filePayload")
	})
}

func TestParseCommandLineArgumentsWithConnectTimeoutEqualToZero(t *testing.T) {
	assert.Panics(t, func() {
		assertConnectTimeout(time.Duration(0))
	})
}

func TestParseCommandLineArgumentsWithConnectTimeout(t *testing.T) {
	assert.NotPanics(t, func() {
		assertConnectTimeout(time.Duration(1))
	})
}

func TestParseCommandLineArgumentsWithRequestsPerSecondLessThanZero(t *testing.T) {
	assert.Panics(t, func() {
		assertRequestsPerSecond(-1)
	})
}

func TestParseCommandLineArgumentsWithLoadDurationZero(t *testing.T) {
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

func TestParseCommandLineArgumentsWithRequestsPerSecond(t *testing.T) {
	assert.NotPanics(t, func() {
		assertRequestsPerSecond(1)
	})
}

func TestParseCommandLineArgumentsWithTotalRequestsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0, 1, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithConcurrencyMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 0, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithConnectionsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 1, 0,
		)
	})
}

func TestParseCommandLineArgumentsWithTotalRequestsMustBeGreaterThanOrEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 2, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithTotalRequestsIsEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			2, 2, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithTotalRequestsIsGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			4, 2, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithConnectionsMustNotBeGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			10, 5, 10,
		)
	})
}

func TestParseCommandLineArgumentsWithConcurrencyMustBeAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			10, 5, 8,
		)
	})
}

func TestParseCommandLineArgumentsWithConcurrencyIsAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			100, 10, 5,
		)
	})
}

func TestParseCommandLineArgumentsWithConcurrencyMustBeGreaterThanZeroEvenIfLoadDurationIsGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1, 0, 1,
		)
	})
}

func TestParseCommandLineArgumentsWithMaxProcsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertAndSetMaxProcs(0)
	})
}

func TestParseCommandLineArgumentsWithMaxProcsIsGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertAndSetMaxProcs(1)
	})
}

func TestParseCommandLineArgumentsWithResponseSizeLessThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertResponseReading(true, -1, 10, 10)
	})
}

func TestParseCommandLineArgumentsWithBothTotalResponsesAndSuccessfulResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertResponseReading(true, 100, 10, 10)
	})
}

func TestParseCommandLineArgumentsWithOnlyTotalResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 10, 0)
	})
}

func TestParseCommandLineArgumentsWithOnlySuccessfulResponsesSpecified(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 0, 10)
	})
}

func TestParseCommandLineArgumentsWithoutResponseReading(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertResponseReading(false, 100, 10, 10)
	})
}

func TestParseCommandLineArgumentsWithNonExistingFile(t *testing.T) {
	assert.Panics(t, func() {
		getFilePayload("./non-existing")
	})
}

func TestParseCommandLineArgumentsWithAnExistingFile(t *testing.T) {
	file, err := os.Create("testFile")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(file.Name())
	}()

	assert.NotPanics(t, func() {
		getFilePayload("./testFile")
	})
}
