package metrics

import (
	"expvar"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestHistogram(t *testing.T) {
	h := NewExpvarHistogram("test_histogram", 0, 100, 3)

	rand.Seed(42)
	const mean, stdev int64 = 50, 10
	for i := 0; i < 1234; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}

	var tolerance int64 = 2
	for quantile, want := range map[string]int64{
		"_p50": 50, // TODO
		"_p90": 61,
		"_p95": 65,
		"_p99": 71,
	} {
		s := expvar.Get("test_histogram" + quantile).String()

		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}

		if int64(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("%s: want %d, have %d", quantile, want, have)
		}
	}
}
