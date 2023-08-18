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
    SuccessCount: 999
    ErrorCount: 1
    TotalPayloadSize: 2.0 kB
    AveragePayloadSize: 20 B
    EarliestLoadSendTime: August 21, 2023 04:14:00 IST
    LatestLoadSendTime: August 21, 2023 04:14:10 IST
    TimeToCompleteLoad: 10s

  Error distribution:
  [1]   load error

  
  ResponseMetrics:
    SuccessCount: 1000
    ErrorCount: 1
    TotalResponsePayloadSize: 1.8 kB
    AverageResponsePayloadSize: 18 B 
    EarliestResponseReceivedTime: August 21, 2023 04:14:00 IST
    LatestResponseReceivedTime: August 21, 2023 04:14:10 IST
    TimeToGetResponses: 10s
  
  Error distribution: 
  [1]   response error

`
	startTime, err := time.Parse(timeFormat, "August 21, 2023 04:14:00 IST")
	assert.Nil(t, err)

	tenSecondsLater, err := time.Parse(timeFormat, "August 21, 2023 04:14:10 IST")
	assert.Nil(t, err)

	report := &Report{
		Load: LoadMetrics{
			TotalRequests:             1000,
			SuccessCount:              999,
			ErrorCount:                1,
			ErrorCountByType:          map[string]uint{"load error": 1},
			TotalPayloadLengthBytes:   2000,
			AveragePayloadLengthBytes: 20.0,
			EarliestLoadSendTime:      startTime,
			LatestLoadSendTime:        tenSecondsLater,
			TotalTime:                 tenSecondsLater.Sub(startTime),
		},
		Response: ResponseMetrics{
			SuccessCount:                      1000,
			ErrorCount:                        1,
			ErrorCountByType:                  map[string]uint{"response error": 1},
			TotalResponsePayloadLengthBytes:   1800,
			AverageResponsePayloadLengthBytes: 18.0,
			EarliestResponseReceivedTime:      startTime,
			LatestResponseReceivedTime:        tenSecondsLater,
			TotalTime:                         tenSecondsLater.Sub(startTime),
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
    SuccessCount: 999
    ErrorCount: 1
    TotalPayloadSize: 2.0 kB
    AveragePayloadSize: 20 B
    EarliestLoadSendTime: August 21, 2023 04:14:00 IST
    LatestLoadSendTime: August 21, 2023 04:14:00 IST
    TimeToCompleteLoad: 0s

  Error distribution:
  [1]   load error


`
	time, err := time.Parse(timeFormat, "August 21, 2023 04:14:00 IST")
	assert.Nil(t, err)

	report := &Report{
		Load: LoadMetrics{
			TotalRequests:             1000,
			SuccessCount:              999,
			ErrorCount:                1,
			ErrorCountByType:          map[string]uint{"load error": 1},
			TotalPayloadLengthBytes:   2000,
			AveragePayloadLengthBytes: 20.0,
			EarliestLoadSendTime:      time,
			LatestLoadSendTime:        time,
			TotalTime:                 time.Sub(time),
		},
		Response: ResponseMetrics{
			IsAvailableForReporting: false,
		},
	}

	buffer := &bytes.Buffer{}
	err = print(buffer, report)

	assert.Equal(t, strings.Trim(expected, " "), strings.Trim(string(buffer.Bytes()), " "))
}
