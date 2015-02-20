package metrics

import (
	"expvar"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestHistogram(t *testing.T) {
	quantiles := []int{50, 90, 95, 99}
	h := NewExpvarHistogram("test_histogram", 0, 100, 3, quantiles...)

	rand.Seed(42)
	const mean, stdev int64 = 50, 10
	for i := 0; i < 1234; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}

	var tolerance int64 = 2
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)
		s := expvar.Get(fmt.Sprintf("test_histogram_p%02d", quantile)).String()
		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		if int64(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}

// https://en.wikipedia.org/wiki/Normal_distribution#Quantile_function
func normalValueAtQuantile(mean, stdev int64, quantile int) int64 {
	return int64(float64(mean) + float64(stdev)*math.Sqrt2*erfinv(2*(float64(quantile)/100)-1))
}

// https://stackoverflow.com/questions/5971830/need-code-for-inverse-error-function
func erfinv(y float64) float64 {
	if y < -1.0 || y > 1.0 {
		panic("invalid input")
	}

	var (
		a = [4]float64{0.886226899, -1.645349621, 0.914624893, -0.140543331}
		b = [4]float64{-2.118377725, 1.442710462, -0.329097515, 0.012229801}
		c = [4]float64{-1.970840454, -1.624906493, 3.429567803, 1.641345311}
		d = [2]float64{3.543889200, 1.637067800}
	)

	const y0 = 0.7
	var x, z float64

	if math.Abs(y) == 1.0 {
		x = -y * math.Log(0.0)
	} else if y < -y0 {
		z = math.Sqrt(-math.Log((1.0 + y) / 2.0))
		x = -(((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
	} else {
		if y < y0 {
			z = y * y
			x = y * (((a[3]*z+a[2])*z+a[1])*z + a[0]) / ((((b[3]*z+b[3])*z+b[1])*z+b[0])*z + 1.0)
		} else {
			z = math.Sqrt(-math.Log((1.0 - y) / 2.0))
			x = (((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
		}
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
	}

	return x
}
