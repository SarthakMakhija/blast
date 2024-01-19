package workers

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
)

func TestNewRequestId(t *testing.T) {
	requestId := NewRequestId()
	assert.Equal(t, uint64(1), requestId.Next())
}

func TestNewRequestIdConcurrently(t *testing.T) {
	var requestIds []uint64
	var lock sync.Mutex

	var wg sync.WaitGroup
	wg.Add(100)

	requestId := NewRequestId()
	for goroutineId := 1; goroutineId <= 100; goroutineId++ {
		go func() {
			defer func() {
				lock.Unlock()
				wg.Done()
			}()

			lock.Lock()
			requestIds = append(requestIds, requestId.Next())
		}()
	}
	wg.Wait()

	sort.Slice(requestIds, func(i, j int) bool {
		return requestIds[i] < requestIds[j]
	})

	var expectedRequestIds []uint64
	for id := uint64(1); id <= 100; id++ {
		expectedRequestIds = append(expectedRequestIds, id)
	}

	assert.Equal(t, expectedRequestIds, requestIds)
}
