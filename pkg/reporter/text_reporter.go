package reporter

import (
	"fmt"
	"io"
)

func PrintTextMetrics(w io.Writer, m Metrics) {
	fmt.Fprintf(w, "%-30s: %d\n", "Total Requests", m.TotalRequests)
	fmt.Fprintf(w, "%-30s: %s\n", "Total Duration", m.TotalDuration)
	fmt.Fprintf(w, "%-30s: %s\n", "Average Response Time", m.AverageResponseTime)
	fmt.Fprintf(w, "%-30s: %s\n", "Minimum Response Time", m.MinResponseTime)
	fmt.Fprintf(w, "%-30s: %s\n", "Maximum Response Time", m.MaxResponseTime)
	fmt.Fprintf(w, "%-30s: %s\n", "Median Response Time", m.MedianResponseTime)
	fmt.Fprintf(w, "%-30s: %s\n", "95th Percentile Response Time", m.Percentile95Time)
	fmt.Fprintf(w, "%-30s: %.2f\n", "Requests Per Second", m.RequestsPerSecond)
}
