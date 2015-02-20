package metrics

import (
	"expvar"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
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
// The histogram exposes metrics for 50th, 90th, 95th, and 99th quantiles as
// gauges. The names are assigned by using the passed name as a prefix and
// appending e.g. "_p50".
func NewExpvarHistogram(name string, minValue, maxValue int64, sigfigs int) Histogram {
	h := &expvarHistogram{
		name: name,
		hist: hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
		p50:  NewExpvarGauge(name + "_p50"),
		p90:  NewExpvarGauge(name + "_p90"),
		p95:  NewExpvarGauge(name + "_p95"),
		p99:  NewExpvarGauge(name + "_p99"),
	}
	go h.rotateLoop(1 * time.Minute)
	return h
}

type expvarHistogram struct {
	name string

	mu   sync.Mutex
	hist *hdrhistogram.WindowedHistogram

	p50 Gauge
	p90 Gauge
	p95 Gauge
	p99 Gauge
}

func (h *expvarHistogram) With(...Field) Histogram { return h }

func (h *expvarHistogram) Observe(value int64) {
	h.mu.Lock()
	err := h.hist.Current.RecordValue(value)
	h.mu.Unlock()

	if err != nil {
		panic(err.Error())
	}

	h.p50.Set(h.hist.Current.ValueAtQuantile(50))
	h.p90.Set(h.hist.Current.ValueAtQuantile(90))
	h.p95.Set(h.hist.Current.ValueAtQuantile(95))
	h.p99.Set(h.hist.Current.ValueAtQuantile(99))
}

func (h *expvarHistogram) rotateLoop(d time.Duration) {
	for _ = range time.Tick(d) {
		h.mu.Lock()
		h.hist.Rotate()
		h.mu.Unlock()
	}
}
