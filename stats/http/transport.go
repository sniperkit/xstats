package httpstats

import (
	// "io"
	"net/http"
	"time"

	"github.com/sniperkit/logger"
	"github.com/sniperkit/stats"
)

// NewTransport wraps t to produce metrics on the default engine for every request sent and every response received.
func NewTransport(t http.RoundTripper) http.RoundTripper {
	stats.Log.Entry.InfoWithFields(logger.Fields{
		"http.RoundTripper": t != nil,
	}, "stats.collector.httpstats.NewTransport()")
	return NewTransportWith(stats.DefaultEngine, t)
}

// NewTransportWith wraps t to produce metrics on eng for every request sent and
// every response received.
func NewTransportWith(eng *stats.Engine, t http.RoundTripper) http.RoundTripper {
	stats.Log.Entry.InfoWithFields(logger.Fields{
		"http.RoundTripper": t != nil,
		"stats.Engine":      eng != nil,
	}, "stats.collector.httpstats.NewTransportWith()")
	return &transport{
		transport: t,
		eng:       eng,
	}
}

type transport struct {
	transport http.RoundTripper
	eng       *stats.Engine
}

func (t *transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	start := time.Now()
	rtrip := t.transport

	stats.Log.Entry.WarningWithFields(logger.Fields{"rtrip": rtrip == nil}, "httpstats.t.RoundTrip()")

	if rtrip == nil {
		rtrip = http.DefaultTransport
	}

	if req.Body == nil {
		req.Body = &nullBody{}
	}

	m := &metrics{}

	req.Body = &requestBody{
		eng:     t.eng,
		req:     req,
		metrics: m,
		body:    req.Body,
		op:      "write",
	}

	// stats.Log.Entry.FatalWithFields(logger.Fields{"req.Body.Header": req.Header}, "httpstats.t.RoundTrip()")

	res, err = rtrip.RoundTrip(req)
	req.Body.Close() // safe guard, the transport should have done it already

	if err != nil {
		stats.Log.Entry.FatalWithFields(logger.Fields{"r.body.Close()": err}, "httpstats.t.RoundTrip()")
		m.observeError(time.Now().Sub(start))
		t.eng.ReportAt(start, m)
	} else {

		stats.Log.Entry.WarningWithFields(logger.Fields{
			"start":   start,
			"eng":     t.eng,
			"res":     res,
			"metrics": m,
			"body":    res.Body,
			"op":      "read",
		}, "stats.collector.httpstats.t.RoundTrip()")

		res.Body = &responseBody{
			debug:   true,
			eng:     t.eng,
			res:     res,
			metrics: m,
			body:    res.Body,
			op:      "read",
			start:   start,
		}
	}

	updateStatistics(start, time.Now(), m)

	return
}
