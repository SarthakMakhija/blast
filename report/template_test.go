package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintsTheReportWithLoadAndResponseMetrics(t *testing.T) {
	expected := `
Summary:
  LoadMetrics:
    TotalRequests: 1000
    RequestsPerSecond: 55.3200
    SuccessCount: 999
    ErrorCount: 1
    TotalPayloadSize: 2000 bytes
    AveragePayloadSize: 20.0000 bytes
    EarliestLoadSendTime: August 21, 2023 04:14:00 IST
    LatestLoadSendTime: August 21, 2023 04:14:00 IST

  Error distribution:
  [1]   load error

  
  ResponseMetrics:
    SuccessCount: 1000
    ErrorCount: 1
    TotalResponsePayloadSize: 1800 bytes
    AverageResponsePayloadSize: 18.0000 bytes
    EarliestResponseReceivedTime: August 21, 2023 04:14:00 IST
    LatestResponseReceivedTime: August 21, 2023 04:14:00 IST
  
  Error distribution: 
  [1]   response error

`
	time, err := time.Parse(timeFormat, "August 21, 2023 04:14:00 IST")
	assert.Nil(t, err)

	report := &Report{
		Load: LoadMetrics{
			TotalRequests:             1000,
			RequestsPerSecond:         55.32,
			SuccessCount:              999,
			ErrorCount:                1,
			ErrorCountByType:          map[string]uint{"load error": 1},
			TotalPayloadLengthBytes:   2000,
			AveragePayloadLengthBytes: 20.0,
			EarliestLoadSendTime:      time,
			LatestLoadSendTime:        time,
		},
		Response: ResponseMetrics{
			SuccessCount:                      1000,
			ErrorCount:                        1,
			ErrorCountByType:                  map[string]uint{"response error": 1},
			TotalResponsePayloadLengthBytes:   1800,
			AverageResponsePayloadLengthBytes: 18.0,
			EarliestResponseReceivedTime:      time,
			LatestResponseReceivedTime:        time,
			IsAvailableForReporting:           true,
		},
	}

	buffer := &bytes.Buffer{}
	err = print(buffer, report)

	assert.Equal(t, strings.Trim(expected, " "), strings.Trim(string(buffer.Bytes()), " "))
}

func TestPrintsTheReportWithLoadMetrics(t *testing.T) {
	expected := `
Summary:
  LoadMetrics:
    TotalRequests: 1000
    RequestsPerSecond: 55.3200
    SuccessCount: 999
    ErrorCount: 1
    TotalPayloadSize: 2000 bytes
    AveragePayloadSize: 20.0000 bytes
    EarliestLoadSendTime: August 21, 2023 04:14:00 IST
    LatestLoadSendTime: August 21, 2023 04:14:00 IST

  Error distribution:
  [1]   load error


`
	time, err := time.Parse(timeFormat, "August 21, 2023 04:14:00 IST")
	assert.Nil(t, err)

	report := &Report{
		Load: LoadMetrics{
			TotalRequests:             1000,
			RequestsPerSecond:         55.32,
			SuccessCount:              999,
			ErrorCount:                1,
			ErrorCountByType:          map[string]uint{"load error": 1},
			TotalPayloadLengthBytes:   2000,
			AveragePayloadLengthBytes: 20.0,
			EarliestLoadSendTime:      time,
			LatestLoadSendTime:        time,
		},
		Response: ResponseMetrics{
			IsAvailableForReporting: false,
		},
	}

	buffer := &bytes.Buffer{}
	err = print(buffer, report)

	assert.Equal(t, strings.Trim(expected, " "), strings.Trim(string(buffer.Bytes()), " "))
}
