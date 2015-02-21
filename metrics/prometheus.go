package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus has strong opinions about the dimensionality of fields. Users
// must predeclare every field key they intend to use. On every observation,
// fields with keys that haven't been predeclared will be silently dropped,
// and predeclared field keys without values will receive the value
// PrometheusLabelValueUnknown.
var PrometheusLabelValueUnknown = "unknown"

type prometheusCounter struct {
	*prometheus.CounterVec
	Pairs map[string]string
}

// NewPrometheusCounter returns a new Counter backed by a Prometheus metric.
// The counter is automatically registered via prometheus.Register.
func NewPrometheusCounter(name, help string, fieldKeys []string) Counter {
	return NewPrometheusCounterWithLabels(name, help, fieldKeys, prometheus.Labels{})
}

// NewPrometheusCounterWithLabels is the same as NewPrometheusCounter, but
// attaches a set of const label pairs to the metric.
func NewPrometheusCounterWithLabels(name, help string, fieldKeys []string, constLabels prometheus.Labels) Counter {
	m := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	p := map[string]string{}
	for _, fieldName := range fieldKeys {
		p[fieldName] = PrometheusLabelValueUnknown
	}

	return prometheusCounter{
		CounterVec: m,
		Pairs:      p,
	}
}

func (c prometheusCounter) With(f Field) Counter {
	return prometheusCounter{
		CounterVec: c.CounterVec,
		Pairs:      merge(c.Pairs, f),
	}
}

func (c prometheusCounter) Add(delta uint64) {
	c.CounterVec.With(prometheus.Labels(c.Pairs)).Add(float64(delta))
}

type prometheusGauge struct {
	*prometheus.GaugeVec
	Pairs map[string]string
}

// NewPrometheusGauge returns a new Gauge backed by a Prometheus metric.
// The gauge is automatically registered via prometheus.Register.
func NewPrometheusGauge(name, help string, fieldKeys []string) Gauge {
	return NewPrometheusGaugeWithLabels(name, help, fieldKeys, prometheus.Labels{})
}

// NewPrometheusGaugeWithLabels is the same as NewPrometheusGauge, but
// attaches a set of const label pairs to the metric.
func NewPrometheusGaugeWithLabels(name, help string, fieldKeys []string, constLabels prometheus.Labels) Gauge {
	m := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	return prometheusGauge{
		GaugeVec: m,
		Pairs:    pairsFrom(fieldKeys),
	}
}

func (g prometheusGauge) With(f Field) Gauge {
	return prometheusGauge{
		GaugeVec: g.GaugeVec,
		Pairs:    merge(g.Pairs, f),
	}
}

func (g prometheusGauge) Set(value int64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Set(float64(value))
}

func (g prometheusGauge) Add(delta int64) {
	g.GaugeVec.With(prometheus.Labels(g.Pairs)).Add(float64(delta))
}

type prometheusHistogram struct {
	*prometheus.SummaryVec
	Pairs map[string]string
}

// NewPrometheusHistogram returns a new Histogram backed by a Prometheus
// summary. It uses a 10-second max age for bucketing. The histogram is
// automatically registered via prometheus.Register.
func NewPrometheusHistogram(name, help string, fieldKeys []string) Histogram {
	return NewPrometheusHistogramWithLabels(name, help, fieldKeys, prometheus.Labels{})
}

// NewPrometheusHistogramWithLabels is the same as NewPrometheusHistogram, but
// attaches a set of const label pairs to the metric.
func NewPrometheusHistogramWithLabels(name, help string, fieldKeys []string, constLabels prometheus.Labels) Histogram {
	m := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:        name,
			Help:        help,
			ConstLabels: constLabels,
			MaxAge:      10 * time.Second,
		},
		fieldKeys,
	)
	prometheus.MustRegister(m)

	return prometheusHistogram{
		SummaryVec: m,
		Pairs:      pairsFrom(fieldKeys),
	}
}

func (h prometheusHistogram) With(f Field) Histogram {
	return prometheusHistogram{
		SummaryVec: h.SummaryVec,
		Pairs:      merge(h.Pairs, f),
	}
}

func (h prometheusHistogram) Observe(value int64) {
	h.SummaryVec.With(prometheus.Labels(h.Pairs)).Observe(float64(value))
}

func pairsFrom(fieldKeys []string) map[string]string {
	p := map[string]string{}
	for _, fieldName := range fieldKeys {
		p[fieldName] = PrometheusLabelValueUnknown
	}
	return p
}

func merge(orig map[string]string, f Field) map[string]string {
	if _, ok := orig[f.Key]; !ok {
		return orig
	}

	newPairs := make(map[string]string, len(orig))
	for k, v := range orig {
		newPairs[k] = v
	}

	newPairs[f.Key] = f.Value
	return newPairs
}
