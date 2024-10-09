package reporter

import (
	"log/slog"
	"sort"
	"time"

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

	var minDuration time.Duration = pairs[0].EndTime.Sub(pairs[0].StartTime)
	var maxDuration time.Duration = minDuration
	durations := make([]time.Duration, 0, totalRequests)
	durationsTotal := 0

	for _, pair := range pairs {
		duration := pair.EndTime.Sub(pair.StartTime)
		durations = append(durations, duration)

		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}

		durationsTotal += int(duration.Seconds())
	}

	// Sort durations to calculate median and percentile
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// Calculate median
	medianResponseTime := durations[totalRequests/2]
	if totalRequests%2 == 0 {
		medianResponseTime = (durations[totalRequests/2-1] + durations[totalRequests/2]) / 2
	}

	// Calculate 95th percentile (rounded down)
	index95 := int(float64(totalRequests) * 0.95)
	percentile95Time := durations[index95]

	// Calculate average response time
	averageResponseTime := time.Duration(durationsTotal/totalRequests) * time.Second

	// Calculate requests per second
	totalTimeFrame := pairs[totalRequests-1].EndTime.Sub(pairs[0].StartTime)
	requestsPerSecond := float64(totalRequests) / totalTimeFrame.Seconds()

	return Metrics{
		TotalRequests:       totalRequests,
		TotalDuration:       totalDuration,
		AverageResponseTime: averageResponseTime,
		MinResponseTime:     minDuration,
		MaxResponseTime:     maxDuration,
		MedianResponseTime:  medianResponseTime,
		Percentile95Time:    percentile95Time,
		RequestsPerSecond:   requestsPerSecond,
	}
}
