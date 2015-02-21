package metrics_test

import (
	"expvar"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/peterbourgon/gokit/metrics"
)

func TestExpvarHistogramQuantiles(t *testing.T) {
	metricName := "test_histogram"
	quantiles := []int{50, 90, 95, 99}
	h := metrics.NewExpvarHistogram(metricName, 0, 100, 3, quantiles...)

	const seed, mean, stdev int64 = 424242, 50, 10
	populateNormalHistogram(t, h, seed, mean, stdev)
	assertExpvarNormalHistogram(t, metricName, mean, stdev, quantiles)
}

func assertExpvarNormalHistogram(t *testing.T, metricName string, mean, stdev int64, quantiles []int) {
	const tolerance int = 2
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)
		s := expvar.Get(fmt.Sprintf("%s_p%02d", metricName, quantile)).String()
		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}
