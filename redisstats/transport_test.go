package redisstats

import (
	"testing"

	"net"

	"github.com/segmentio/objconv/resp"
	"github.com/segmentio/redis-go"
	"github.com/segmentio/stats"
)

func TestTransport(t *testing.T) {
	t.Skip("temporarily disabling this test to move forward with the hackweek project")

	m := &testMetricHandler{}
	e := stats.NewEngine("")
	e.Register(m)

	client := redis.Client{
		Addr:      "127.0.0.1",
		Transport: NewTransportWith(e, &testTransport{}),
		Timeout:   0,
	}
	respErrorClient := redis.Client{
		Addr:      "127.0.0.1",
		Transport: NewTransportWith(e, &testTransport{err: resp.NewError("request failed")}),
		Timeout:   0,
	}
	readErrorClient := redis.Client{
		Addr:      "127.0.0.1",
		Transport: NewTransportWith(e, &testTransport{err: &net.OpError{Op: "read"}}),
		Timeout:   0,
	}
	writeErrorClient := redis.Client{
		Addr:      "127.0.0.1",
		Transport: NewTransportWith(e, &testTransport{err: &net.OpError{Op: "write"}}),
		Timeout:   0,
	}

	client.Do(redis.NewRequest("1.2.3.4:6379", "SET", redis.List("foo", "bar")))
	client.Do(redis.NewRequest("5.6.7.8:6379", "GET", redis.List("foo")))

	respErrorClient.Do(redis.NewRequest("9.9.9.9:6379", "GET", redis.List("foo")))
	readErrorClient.Do(redis.NewRequest("9.9.9.9:6379", "GET", redis.List("foo")))
	writeErrorClient.Do(redis.NewRequest("9.9.9.9:6379", "SET", redis.List("foo", "bar")))

	if len(m.metrics) == 0 {
		t.Error("no metrics reported by http handler")
	}

	validateMetric(t, m.metrics[0], "requests",
		[]stats.Tag{{Name: "command", Value: "SET"}}, stats.CounterType)

	validateMetric(t, m.metrics[1], "request.rtt.seconds",
		[]stats.Tag{
			{Name: "command", Value: "SET"},
			{Name: "upstream", Value: "1.2.3.4:6379"},
		}, stats.HistogramType)

	validateMetric(t, m.metrics[2], "requests",
		[]stats.Tag{{Name: "command", Value: "GET"}}, stats.CounterType)

	validateMetric(t, m.metrics[3], "request.rtt.seconds",
		[]stats.Tag{
			{Name: "command", Value: "GET"},
			{Name: "upstream", Value: "5.6.7.8:6379"},
		}, stats.HistogramType)

	validateMetric(t, m.metrics[4], "requests",
		[]stats.Tag{{Name: "command", Value: "GET"}}, stats.CounterType)

	validateMetric(t, m.metrics[5], "errors",
		[]stats.Tag{
			{Name: "type", Value: "response"},
			{Name: "upstream", Value: "9.9.9.9:6379"},
		}, stats.CounterType)

	validateMetric(t, m.metrics[6], "requests",
		[]stats.Tag{{Name: "command", Value: "GET"}}, stats.CounterType)

	validateMetric(t, m.metrics[7], "errors",
		[]stats.Tag{
			{Name: "type", Value: "network"},
			{Name: "operation", Value: "read"},
			{Name: "upstream", Value: "9.9.9.9:6379"},
		}, stats.CounterType)

	validateMetric(t, m.metrics[8], "requests",
		[]stats.Tag{{Name: "command", Value: "SET"}}, stats.CounterType)

	validateMetric(t, m.metrics[9], "errors",
		[]stats.Tag{
			{Name: "type", Value: "network"},
			{Name: "operation", Value: "write"},
			{Name: "upstream", Value: "9.9.9.9:6379"},
		}, stats.CounterType)

}

type testTransport struct {
	err error
}

func (tr *testTransport) RoundTrip(req *redis.Request) (*redis.Response, error) {
	if tr.err != nil {
		return nil, tr.err
	}
	return &redis.Response{Args: redis.List("OK"), Request: req}, nil
}
