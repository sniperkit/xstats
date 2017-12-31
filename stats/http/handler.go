package httpstats

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/sniperkit/xlogger/pkg"
	"github.com/sniperkit/xstats/pkg"
)

// NewHandler wraps h to produce metrics on the default engine for every request
// received and every response sent.
func NewHandler(h http.Handler) http.Handler {
	stats.Log.Entry.InfoWithFields(logger.Fields{
		"http.Handler": h != nil,
	}, "stats.collector.httpstats.NewHandler()")
	return NewHandlerWith(stats.DefaultEngine, h)
}

// NewHandlerWith wraps h to produce metrics on eng for every request received
// and every response sent.
func NewHandlerWith(eng *stats.Engine, h http.Handler) http.Handler {
	stats.Log.Entry.InfoWithFields(logger.Fields{
		"http.Handler": h != nil,
		"stats.Engine": eng != nil,
	}, "stats.collector.httpstats.NewHandlerWith()")
	return &handler{
		handler: h,
		eng:     eng,
	}
}

type handler struct {
	handler http.Handler
	eng     *stats.Engine
}

func (h *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m := &metrics{}

	w := &responseWriter{
		ResponseWriter: res,
		eng:            h.eng,
		req:            req,
		metrics:        m,
		start:          time.Now(),
	}
	defer w.complete()

	b := &requestBody{
		body:    req.Body,
		eng:     h.eng,
		req:     req,
		metrics: m,
		op:      "read",
	}
	defer b.close()
	stats.Log.Entry.InfoWithFields(logger.Fields{
		"requestBody":    b != nil,
		"responseWriter": w != nil,
	}, "stats.collector.httpstats.handler.ServeHTTP()")

	req.Body = b
	h.handler.ServeHTTP(w, req)
}

type responseWriter struct {
	http.ResponseWriter
	start       time.Time
	eng         *stats.Engine
	req         *http.Request
	metrics     *metrics
	status      int
	bytes       int
	wroteHeader bool
	wroteStats  bool
}

func (w *responseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = http.StatusOK
	}

	if n, err = w.ResponseWriter.Write(b); n > 0 {
		w.bytes += n
	}

	return
}

func (w *responseWriter) Hijack() (conn net.Conn, buf *bufio.ReadWriter, err error) {
	if conn, buf, err = w.ResponseWriter.(http.Hijacker).Hijack(); err == nil {
		w.wroteHeader = true
		w.complete()
	}
	return
}

func (w *responseWriter) complete() {
	if w.wroteStats {
		return
	}
	w.wroteStats = true

	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = http.StatusOK
	}

	now := time.Now()
	res := &http.Response{
		ProtoMajor:    w.req.ProtoMajor,
		ProtoMinor:    w.req.ProtoMinor,
		Proto:         w.req.Proto,
		StatusCode:    w.status,
		Header:        w.Header(),
		Request:       w.req,
		ContentLength: -1,
	}

	w.metrics.observeResponse(res, "write", w.bytes, now.Sub(w.start))
	w.eng.ReportAt(w.start, w.metrics)

	stats.Log.Entry.DebugWithFields(logger.Fields{
		"w.metrics": w.metrics,
		"w.start":   w.start,
	}, "stats.collector.httpstats.responseWriter.complete()")

}
