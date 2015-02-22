package metrics_test

import (
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/Godeps/_workspace/src/github.com/prometheus/client_golang/prometheus"
	"github.com/peterbourgon/gokit/metrics"
)

func TestPrometheusLabelBehavior(t *testing.T) {
	c := metrics.NewPrometheusCounter("test_prometheus_label_behavior", "Abc def ghi.", []string{"used_key", "unused_key"})
	c.With(metrics.Field{Key: "used_key", Value: "declared"}).Add(1)
	c.Add(1)

	if want, have := strings.Join([]string{
		`# HELP test_prometheus_label_behavior Abc def ghi.`,
		`# TYPE test_prometheus_label_behavior counter`,
		`test_prometheus_label_behavior{unused_key="unknown",used_key="declared"} 1`,
		`test_prometheus_label_behavior{unused_key="unknown",used_key="unknown"} 1`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusCounter(t *testing.T) {
	c := metrics.NewPrometheusCounter("test_prometheus_counter", "Lorem ipsum.", []string{})
	c.Add(1)
	c.Add(2)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_counter Lorem ipsum.`,
		`# TYPE test_prometheus_counter counter`,
		`test_prometheus_counter 3`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(3)
	c.Add(4)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_counter Lorem ipsum.`,
		`# TYPE test_prometheus_counter counter`,
		`test_prometheus_counter 10`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusGauge(t *testing.T) {
	c := metrics.NewPrometheusGauge("test_gauge", "Dolor sit.", []string{})
	c.Set(42)
	if want, have := strings.Join([]string{
		`# HELP test_gauge Dolor sit.`,
		`# TYPE test_gauge gauge`,
		`test_gauge 42`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(-43)
	if want, have := strings.Join([]string{
		`# HELP test_gauge Dolor sit.`,
		`# TYPE test_gauge gauge`,
		`test_gauge -1`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusHistogram(t *testing.T) {
	h := metrics.NewPrometheusHistogram("test_histogram", "Qwerty asdf.", []string{})

	const mean, stdev int64 = 50, 10
	populateNormalHistogram(t, h, 34, mean, stdev)
	assertPrometheusNormalHistogram(t, "test_histogram", mean, stdev)
}

func assertPrometheusNormalHistogram(t *testing.T, metricName string, mean, stdev int64) {
	scrape := scrapePrometheus(t)
	const tolerance int = 5 // Prometheus approximates higher quantiles badly -_-;
	for quantileInt, quantileStr := range map[int]string{50: "0.5", 90: "0.9", 99: "0.99"} {
		want := normalValueAtQuantile(mean, stdev, quantileInt)
		have := getPrometheusQuantile(t, scrape, metricName, quantileStr)
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("%q: want %d, have %d", quantileStr, want, have)
		}
	}
}

func scrapePrometheus(t *testing.T) string {
	server := httptest.NewServer(prometheus.UninstrumentedHandler())
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(buf))
}

func getPrometheusQuantile(t *testing.T, scrape, name, quantileStr string) int {
	matches := regexp.MustCompile(name+`{quantile="`+quantileStr+`"} ([0-9]+)`).FindAllStringSubmatch(scrape, -1)
	if len(matches) < 1 {
		t.Fatalf("%q: quantile %q not found in scrape", name, quantileStr)
	}
	if len(matches[0]) < 2 {
		t.Fatalf("%q: quantile %q not found in scrape", name, quantileStr)
	}
	i, err := strconv.Atoi(matches[0][1])
	if err != nil {
		t.Fatal(err)
	}
	return i
}
