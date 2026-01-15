package reporter

import (
	"log/slog"
	"time"

	"github.com/jamiealquiza/tachymeter"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/tracker"
)

type Metrics struct {
	TotalRequests       int
	TotalDuration       time.Duration
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	MedianResponseTime  time.Duration
	Percentile95Time    time.Duration
	RequestsPerSecond   float64
}

// CalculateMetrics calculates the desired metrics from an array of RequestTrackerPair
func CalculateMetrics(pairs []tracker.RequestTrackerPair, totalDuration time.Duration) Metrics {
	totalRequests := len(pairs)
	slog.Debug("Total Requests: ", "req", totalRequests)
	if totalRequests == 0 {
		return Metrics{}
	}

	t := tachymeter.New(&tachymeter.Config{Size: totalRequests})

	for _, pair := range pairs {
		t.AddTime(pair.EndTime.Sub(pair.StartTime))
	}

	results := t.Calc()

	return Metrics{
		TotalRequests:       totalRequests,
		TotalDuration:       totalDuration,
		AverageResponseTime: results.Time.Avg,
		MinResponseTime:     results.Time.Min,
		MaxResponseTime:     results.Time.Max,
		MedianResponseTime:  results.Time.P50,
		Percentile95Time:    results.Time.P95,
		RequestsPerSecond:   results.Rate.Second,
	}
}
