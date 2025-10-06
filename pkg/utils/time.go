package utils

import (
	"time"
)

type TimeUtils struct{}

func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

func (tu *TimeUtils) Now() time.Time {
	return time.Now().UTC()
}

func (tu *TimeUtils) ParseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return duration
}

func (tu *TimeUtils) FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.String()
	}
	
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	
	return d.Round(time.Second).String()
}

func (tu *TimeUtils) CalculateAge(t time.Time) time.Duration {
	return time.Since(t)
}

func (tu *TimeUtils) IsWithinRange(t, start, end time.Time) bool {
	return !t.Before(start) && !t.After(end)
}

func (tu *TimeUtils) CalculateElapsed(start time.Time) time.Duration {
	return time.Since(start)
}

func (tu *TimeUtils) CalculateRemaining(start time.Time, total time.Duration) time.Duration {
	elapsed := time.Since(start)
	if elapsed >= total {
		return 0
	}
	return total - elapsed
}

func (tu *TimeUtils) RoundToNearest(t time.Time, d time.Duration) time.Time {
	if d <= 0 {
		return t
	}
	
	round := t.Round(d)
	if round.After(t) {
		return round.Add(-d)
	}
	return round
}

func (tu *TimeUtils) CeilToNearest(t time.Time, d time.Duration) time.Time {
	if d <= 0 {
		return t
	}
	
	round := t.Round(d)
	if round.Before(t) {
		return round.Add(d)
	}
	return round
}

func (tu *TimeUtils) FloorToNearest(t time.Time, d time.Duration) time.Time {
	if d <= 0 {
		return t
	}
	
	round := t.Round(d)
	if round.After(t) {
		return round.Add(-d)
	}
	return round
}

func (tu *TimeUtils) CalculateAverageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	
	return total / time.Duration(len(durations))
}

func (tu *TimeUtils) CalculateMaxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}

func (tu *TimeUtils) CalculateMinDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	min := durations[0]
	for _, d := range durations {
		if d < min {
			min = d
		}
	}
	return min
}

func (tu *TimeUtils) IsExpired(t time.Time, expiry time.Duration) bool {
	return time.Since(t) > expiry
}

func (tu *TimeUtils) TimeUntil(t time.Time) time.Duration {
	now := time.Now()
	if t.Before(now) {
		return 0
	}
	return t.Sub(now)
}

func (tu *TimeUtils) FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (tu *TimeUtils) ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func (tu *TimeUtils) CalculateTimeDifference(t1, t2 time.Time) time.Duration {
	return t2.Sub(t1)
}

func (tu *TimeUtils) AbsDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func (tu *TimeUtils) CalculateJitter(durations []time.Duration) time.Duration {
	if len(durations) < 2 {
		return 0
	}
	
	average := tu.CalculateAverageDuration(durations)
	var sum time.Duration
	
	for _, d := range durations {
		diff := tu.AbsDuration(d - average)
		sum += diff
	}
	
	return sum / time.Duration(len(durations))
}

func (tu *TimeUtils) CalculateThroughput(count int, duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	return float64(count) / duration.Seconds()
}

func (tu *TimeUtils) CalculateLatencyPercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i