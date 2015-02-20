package metrics

import (
	"expvar"
	"fmt"
	"sync"
	"time"

	"github.com/peterbourgon/gokit/Godeps/_workspace/src/github.com/codahale/hdrhistogram"
)

type expvarCounter struct {
	v *expvar.Int
}

// NewExpvarCounter returns a new Counter backed by an expvar with the given
// name. Fields are ignored.
func NewExpvarCounter(name string) Counter {
	return &expvarCounter{expvar.NewInt(name)}
}

func (c *expvarCounter) With(...Field) Counter { return c }

func (c *expvarCounter) Add(delta uint64) { c.v.Add(int64(delta)) }

type expvarGauge struct {
	v *expvar.Int
}

// NewExpvarGauge returns a new Gauge backed by an expvar with the given name.
// Fields are ignored.
func NewExpvarGauge(name string) Gauge {
	return &expvarGauge{expvar.NewInt(name)}
}

func (g *expvarGauge) With(...Field) Gauge { return g }

func (g *expvarGauge) Add(delta int64) { g.v.Add(delta) }

func (g *expvarGauge) Set(value int64) { g.v.Set(value) }

// NewExpvarHistogram is taken from http://github.com/codahale/metrics. It
// returns a windowed HDR histogram which drops data older than five minutes.
//
// The histogram exposes metrics for each passed quantile as gauges. Quantiles
// should be integers in the range 1..99. The gauge names are assigned by
// using the passed name as a prefix and appending "_pNN" e.g. "_p50".
func NewExpvarHistogram(name string, minValue, maxValue int64, sigfigs int, quantiles ...int) Histogram {
	gauges := map[int]Gauge{}
	for _, quantile := range quantiles {
		if quantile <= 0 || quantile >= 100 {
			panic(fmt.Sprintf("invalid quantile %d", quantile))
		}
		gauges[quantile] = NewExpvarGauge(fmt.Sprintf("%s_p%02d", name, quantile))
	}
	h := &expvarHistogram{
		hist:   hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		name:   name,
		gauges: gauges,
	}
	go h.rotateLoop(1 * time.Minute)
	return h
}

type expvarHistogram struct {
	mu   sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	name   string
	gauges map[int]Gauge
}

func (h *expvarHistogram) With(...Field) Histogram { return h }

func (h *expvarHistogram) Observe(value int64) {
	h.mu.Lock()
	err := h.hist.Current.RecordValue(value)
	h.mu.Unlock()

	if err != nil {
		panic(err.Error())
	}

	for q, gauge := range h.gauges {
		gauge.Set(h.hist.Current.ValueAtQuantile(float64(q)))
	}
}

func (h *expvarHistogram) rotateLoop(d time.Duration) {
	for _ = range time.Tick(d) {
		h.mu.Lock()
		h.hist.Rotate()
		h.mu.Unlock()
	}
}
