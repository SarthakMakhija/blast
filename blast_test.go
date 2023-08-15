package main

import (
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

func TestRunBlastWithoutFilePath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertFilePath("")
	})
}

func TestRunBlastWithEmptyFilePath(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertFilePath(" ")
	})
}

func TestTotalRequestsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			0, 1, 1,
		)
	})
}

func TestConcurrencyMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			1, 0, 1,
		)
	})
}

func TestConnectionsMustBeGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			1, 1, 0,
		)
	})
}

func TestTotalRequestsMustBeGreaterThanOrEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			1, 2, 1,
		)
	})
}

func TestTotalRequestsIsEqualToConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			2, 2, 1,
		)
	})
}

func TestTotalRequestsIsGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			4, 2, 1,
		)
	})
}

func TestConnectionsMustNotBeGreaterThanConcurrency(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			10, 5, 10,
		)
	})
}

func TestConcurrencyMustBeAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			10, 5, 8,
		)
	})
}

func TestConcurrencyIsAMultipleOfConnections(t *testing.T) {
	exitFunction = exitWithPanic
	assert.NotPanics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			0*time.Second,
			100, 10, 5,
		)
	})
}

func TestConcurrencyMustBeGreaterThanZeroEvenIfLoadDurationIsGreaterThanZero(t *testing.T) {
	exitFunction = exitWithPanic
	assert.Panics(t, func() {
		assertTotalConcurrentRequestsWithClientConnections(
			1*time.Second,
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
