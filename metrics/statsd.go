package metrics

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
)

// statsd metrics take considerable influence from
// https://github.com/streadway/handy package statsd.

const maxBufferSize = 1400 // bytes

type statsdCounter chan uint64

// NewStatsdCounter returns a Counter that emits observations in the statsd
// protocol to the passed writer. Observations are buffered for the reporting
// interval or until the buffer exceeds a max packet size, whichever comes
// first. Fields are ignored.
func NewStatsdCounter(w io.Writer, key string, interval time.Duration) Counter {
	c := make(chan uint64)
	go fwd(w, key, interval, mapCounter(c))
	return statsdCounter(c)
}

func (c statsdCounter) With(Field) Counter { return c }

func (c statsdCounter) Add(delta uint64) { c <- delta }

func mapCounter(in chan uint64) chan string {
	out := make(chan string)
	go func() {
		for delta := range in {
			out <- fmt.Sprintf(":%d|c", delta)
		}
	}()
	return out
}

type statsdGauge struct {
	add chan int64
	set chan int64
}

// NewStatsdGauge returns a Gauge that emits values in the statsd protocol to
// the passed writer. Values are buffered for the reporting interval or until
// the buffer exceeds a max packet size, whichever comes first. Fields are
// ignored.
func NewStatsdGauge(w io.Writer, key string, interval time.Duration) Gauge {
	g := &statsdGauge{
		add: make(chan int64),
		set: make(chan int64),
	}
	go fwd(w, key, interval, mapGauge(g.add, g.set))
	return g
}

func (g *statsdGauge) With(Field) Gauge { return g }

func (g *statsdGauge) Add(delta int64) { g.add <- delta }

func (g *statsdGauge) Set(value int64) { g.set <- value }

func mapGauge(add, set chan int64) chan string {
	out := make(chan string)
	go func() {
		var v int64
		for {
			select {
			case delta := <-add:
				v += delta
			case value := <-set:
				v = value
			}
			out <- fmt.Sprintf(":%d|g", v)
		}
	}()
	return out
}

type statsdHistogram chan int64

// NewStatsdHistogram returns a Histogram that emits observations in the
// statsd protocol to the passed writer. Observations are buffered for the
// reporting interval or until the buffer exceeds a max packet size, whichever
// comes first. Fields are ignored.
func NewStatsdHistogram(w io.Writer, key string, interval time.Duration) Histogram {
	c := make(chan int64)
	go fwd(w, key, interval, mapHistogram(c))
	return statsdHistogram(c)
}

func (h statsdHistogram) With(Field) Histogram { return h }

func (h statsdHistogram) Observe(value int64) { h <- value }

func mapHistogram(in chan int64) chan string {
	out := make(chan string)
	go func() {
		for observation := range in {
			out <- fmt.Sprintf(":%d|TODO", observation)
		}
	}()
	return out
}

func fwd(w io.Writer, key string, interval time.Duration, c chan string) {
	buf := &bytes.Buffer{}
	tick := time.Tick(interval)
	for {
		select {
		case s := <-c:
			if buf.Len() == 0 {
				buf.WriteString(key)
			}
			fmt.Fprintf(buf, s)
			if buf.Len() > maxBufferSize {
				flush(w, buf)
			}

		case <-tick:
			flush(w, buf)
		}
	}
}

func flush(w io.Writer, buf *bytes.Buffer) {
	if buf.Len() <= 0 {
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("error: could not write to statsd: %v", err)
	}
	buf.Reset()
}
