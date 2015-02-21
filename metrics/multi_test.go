package metrics_test

import (
	"expvar"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/metrics"
)

func TestMultiCounter(t *testing.T) {
	metrics.NewMultiCounter(
		metrics.NewExpvarCounter("alpha"),
		metrics.NewPrometheusCounter("beta", "Beta counter.", []string{}),
	).Add(123)

	if want, have := "123", expvar.Get("alpha").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}

	if want, have := strings.Join([]string{
		`# HELP beta Beta counter.`,
		`# TYPE beta counter`,
		`beta 123`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiGauge(t *testing.T) {
	g := metrics.NewMultiGauge(
		metrics.NewExpvarGauge("delta"),
		metrics.NewPrometheusGauge("kappa", "Kappa gauge.", []string{}),
	)

	g.Set(34)

	if want, have := "34", expvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP kappa Kappa gauge.`,
		`# TYPE kappa gauge`,
		`kappa 34`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}

	g.Add(-40)

	if want, have := "-6", expvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP kappa Kappa gauge.`,
		`# TYPE kappa gauge`,
		`kappa -6`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiHistogram(t *testing.T) {
	quantiles := []int{50, 90, 99}
	h := metrics.NewMultiHistogram(
		metrics.NewExpvarHistogram("omicron", 0, 100, 3, quantiles...),
		metrics.NewPrometheusHistogram("nu", "Nu histogram.", []string{}),
	)

	const seed, mean, stdev int64 = 123, 50, 10
	populateNormalHistogram(t, h, seed, mean, stdev)
	assertExpvarNormalHistogram(t, "omicron", mean, stdev, quantiles)
	assertPrometheusNormalHistogram(t, "nu", mean, stdev)
}
