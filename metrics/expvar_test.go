package metrics_test

import (
	"expvar"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/peterbourgon/gokit/metrics"
)

func TestExpvarHistogramQuantiles(t *testing.T) {
	quantiles := []int{50, 90, 95, 99}
	h := metrics.NewExpvarHistogram("test_histogram", 0, 100, 3, quantiles...)

	rand.Seed(42)
	const mean, stdev int64 = 50, 10
	for i := 0; i < 1234; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}

	var tolerance int = 2
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)
		s := expvar.Get(fmt.Sprintf("test_histogram_p%02d", quantile)).String()
		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}
