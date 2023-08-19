package report

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
)

var templateText = `
Summary:
  LoadMetrics:
    TotalConnections: {{ formatNumberUint .Load.TotalConnections }}
    TotalRequests: {{ formatNumberUint .Load.TotalRequests }}
    SuccessCount: {{ formatNumberUint .Load.SuccessCount }}
    ErrorCount: {{ formatNumberUint .Load.ErrorCount }}
    TotalPayloadSize: {{ humanizePayloadSize .Load.TotalPayloadLengthBytes }}
    AveragePayloadSize: {{ humanizePayloadSize .Load.AveragePayloadLengthBytes }}
    EarliestLoadSendTime: {{ formatTime .Load.EarliestLoadSendTime}}
    LatestLoadSendTime: {{ formatTime .Load.LatestLoadSendTime}}
    TimeToCompleteLoad: {{ formatDuration .Load.TotalTime }}

{{ if gt (len .Load.ErrorCountByType) 0 }}  Error distribution:{{ range $err, $num := .Load.ErrorCountByType }}
  [{{ $num }}]   {{ $err }}{{ end }}{{ else }}  Error distribution:
  none{{ end }}
{{ if eq (.Response.IsAvailableForReporting) true }}  
  ResponseMetrics:
    TotalResponses: {{ formatNumberUint .Response.TotalResponses }}
    SuccessCount: {{ formatNumberUint .Response.SuccessCount }}
    ErrorCount: {{ formatNumberUint .Response.ErrorCount }}
    TotalResponsePayloadSize: {{ humanizePayloadSize .Response.TotalResponsePayloadLengthBytes }}
    AverageResponsePayloadSize: {{ humanizePayloadSize .Response.AverageResponsePayloadLengthBytes }} 
    EarliestResponseReceivedTime: {{ formatTime .Response.EarliestResponseReceivedTime }}
    LatestResponseReceivedTime: {{ formatTime .Response.LatestResponseReceivedTime }}
    TimeToGetResponses: {{ formatDuration .Response.TotalTime }}
  
{{ if gt (len .Response.ErrorCountByType) 0 }}  Error distribution:{{ range $err, $num := .Response.ErrorCountByType }} 
  [{{ $num }}]   {{ $err }}{{ end }}{{ else }}  Error distribution:
  none{{ end }}{{ end }}
`

var functions = template.FuncMap{
	"formatNumberUint":    formatNumberUint,
	"formatNumberInt64":   formatNumberInt64,
	"formatTime":          formatTime,
	"formatDuration":      formatDuration,
	"humanizePayloadSize": humanizePayloadSize,
}

const timeFormat = "January 02, 2006 15:04:05 MST"

func formatNumberUint(value uint) string {
	return fmt.Sprintf("%d", value)
}

func humanizePayloadSize(size int64) string {
	return humanize.Bytes(uint64(size))
}

func formatNumberInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func formatTime(time time.Time) string {
	return time.Format(timeFormat)
}

func formatDuration(duration time.Duration) string {
	return duration.String()
}

func print(writer io.Writer, report *Report) error {
	return newTemplate().Execute(writer, report)
}

func newTemplate() *template.Template {
	return template.Must(template.New("blast").Funcs(functions).Parse(templateText))
}
