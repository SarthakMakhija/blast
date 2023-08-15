package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var expected = `
Summary:
  LoadMetrics:
    SuccessCount: 1
    ErrorCount: 1
    TotalPayloadSize: 20 bytes
    AveragePayloadSize: 20.0000 bytes
    EarliestLoadSendTime: August 21, 2023 04:14:00 IST
    LatestLoadSendTime: August 21, 2023 04:14:00 IST

  Error distribution:
  [1]   load error

  ResponseMetrics:
    SuccessCount: 1
    ErrorCount: 1
    TotalResponsePayloadSize: 18 bytes
    AverageResponsePayloadSize: 18.0000 bytes
    EarliestResponseReceivedTime: August 21, 2023 04:14:00 IST
    LatestResponseReceivedTime: August 21, 2023 04:14:00 IST
  
  Error distribution: 
  [1]   response error
`

func TestPrintsTheReport(t *testing.T) {
	time, err := time.Parse(timeFormat, "August 21, 2023 04:14:00 IST")
	assert.Nil(t, err)

	report := &Report{
		Load: LoadMetrics{
			SuccessCount:              1,
			ErrorCount:                1,
			ErrorCountByType:          map[string]uint{"load error": 1},
			TotalPayloadLengthBytes:   20,
			AveragePayloadLengthBytes: 20.0,
			EarliestLoadSendTime:      time,
			LatestLoadSendTime:        time,
		},
		Response: ResponseMetrics{
			SuccessCount:                      1,
			ErrorCount:                        1,
			ErrorCountByType:                  map[string]uint{"response error": 1},
			TotalResponsePayloadLengthBytes:   18,
			AverageResponsePayloadLengthBytes: 18.0,
			EarliestResponseReceivedTime:      time,
			LatestResponseReceivedTime:        time,
		},
	}

	buffer := &bytes.Buffer{}
	err = print(buffer, report)

	assert.Equal(t, strings.Trim(expected, " "), strings.Trim(string(buffer.Bytes()), " "))
}
