package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/opendevops-cn/codo-golang-sdk/consts"
	"github.com/opendevops-cn/codo-golang-sdk/internal/meta"
	"github.com/opendevops-cn/codo-golang-sdk/xnet/xip"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	intranetIps, _ = xip.GetIntranetIpArray()
	intranetIpStr  = strings.Join(intranetIps, ",")
	hostname, _    = os.Hostname()
)

const (
	metricLabelKind      = "kind"
	metricLabelOperation = "operation"
	metricLabelCode      = "code"
	metricLabelReason    = "reason"
)

const (
	DefaultClientSecondsHistogramName = "client_requests_seconds_bucket"
	DefaultClientRequestsCounterName  = "client_requests_code_total"
)

type IClient interface {
	Do(ctx context.Context, request *http.Request, opts ...IDoOptions) (*http.Response, error)
}

type Client struct {
	client *http.Client

	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	// counter: client_requests_code_total{kind, operation, code, reason}
	requests metric.Int64Counter
	// histogram: client_requests_seconds_bucket{kind, operation}
	seconds metric.Float64Histogram
}

type ClientOptions func(client *Client)

func WithTraceProvider(tr trace.TracerProvider) ClientOptions {
	return func(client *Client) {
		client.tracerProvider = tr
	}
}

func WithMeterProvider(mr metric.MeterProvider) ClientOptions {
	return func(client *Client) {
		client.meterProvider = mr
	}
}

func WithClientOptionsTransport(transport http.RoundTripper) ClientOptions {
	return func(client *Client) {
		client.client.Transport = transport
	}
}

func WithClientOptionsCheckRedirect(cf func(req *http.Request, via []*http.Request) error) ClientOptions {
	return func(client *Client) {
		client.client.CheckRedirect = cf
	}
}

func WithClientOptionsTimeout(timeout time.Duration) ClientOptions {
	return func(client *Client) {
		client.client.Timeout = timeout
	}
}

func WithClientOptionsJar(jar http.CookieJar) ClientOptions {
	return func(client *Client) {
		client.client.Jar = jar
	}
}

func NewClient(opts ...ClientOptions) (IClient, error) {
	client := &Client{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  otel.GetMeterProvider(),
	}
	for _, opt := range opts {
		opt(client)
	}

	var err error

	meter := client.meterProvider.Meter("codo/xhttp")

	if client.seconds == nil {
		client.seconds, err = DefaultSecondsHistogram(meter, DefaultClientSecondsHistogramName)
		if err != nil {
			return nil, err
		}
	}

	if client.requests == nil {
		client.requests, err = DefaultRequestsCounter(meter, DefaultClientRequestsCounterName)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (x *Client) Do(ctx context.Context, request *http.Request, opts ...IDoOptions) (*http.Response, error) {
	doOptions := defaultDoOptions()
	for _, opt := range opts {
		opt.apply(&doOptions)
	}

	var (
		code      int
		reason    string
		kind      = "client"
		operation = request.URL.Path
	)
	startTime := time.Now()

	const (
		tracingEventHttpResponseHeaders = "http.response.headers"
		tracingEventHttpResponseBody    = "http.response.body"
	)

	tr := x.tracerProvider.Tracer(
		"codo/xhttp",
		trace.WithInstrumentationVersion(meta.Version),
	)
	ctx, span := tr.Start(ctx, request.URL.String(), trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(commonLabels()...)

	// Inject tracing content into http header.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))

	// Continue client handler executing.
	response, err := x.client.Do(request)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(`%+v`, err))
		return nil, err
	}

	var reqBodyContentBytes []byte
	if response.ContentLength < consts.MegaByte4 {
		reqBodyContentBytes, _ = io.ReadAll(response.Body)
		response.Body = io.NopCloser(bytes.NewReader(reqBodyContentBytes))

		// 解析 reason
		type reasonData struct {
			Msg string `json:"msg"`
		}
		var rd reasonData
		_ = json.Unmarshal(reqBodyContentBytes, &rd)
		reason = rd.Msg
	}

	// Record response status.
	code = response.StatusCode

	// Record metrics.
	x.requests.Add(
		ctx, 1,
		metric.WithAttributes(
			attribute.String(metricLabelKind, kind),
			attribute.String(metricLabelOperation, operation),
			attribute.Int(metricLabelCode, code),
			attribute.String(metricLabelReason, reason),
		),
	)
	x.seconds.Record(
		ctx, time.Since(startTime).Seconds(),
		metric.WithAttributes(
			attribute.String(metricLabelKind, kind),
			attribute.String(metricLabelOperation, operation),
		),
	)

	// 记录 trace
	bs, _ := json.Marshal(headerToMap(response.Header))
	span.AddEvent("http.response", trace.WithAttributes(
		attribute.String(tracingEventHttpResponseHeaders, string(bs)),
		attribute.String(tracingEventHttpResponseBody, strLimit(
			string(reqBodyContentBytes),
			int(doOptions.recordSize),
			"...",
		)),
	))
	return response, nil
}

func ParseProxy(proxy string) (*url.URL, error) {
	if proxy == "" {
		return nil, nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil ||
		(proxyURL.Scheme != "http" &&
			proxyURL.Scheme != "https" &&
			proxyURL.Scheme != "socks5") {
		// proxy was bogus. Try prepending "http://" to it and
		// see if that parses correctly. If not, we fall
		// through and complain about the original one.
		if proxyURL, err := url.Parse("http://" + proxy); err == nil {
			return proxyURL, nil
		}
	}
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
	}
	return proxyURL, nil
}

// headerToMap coverts request headers to map.
func headerToMap(header http.Header) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range header {
		if len(v) > 1 {
			m[k] = v
		} else {
			m[k] = v[0]
		}
	}
	return m
}

// commonLabels returns common used attribute labels:
// ip.intranet, hostname.
func commonLabels() []attribute.KeyValue {
	const (
		tracingCommonKeyIpIntranet = `ip.intranet`
		tracingCommonKeyIpHostname = `hostname`
	)

	return []attribute.KeyValue{
		attribute.String(tracingCommonKeyIpHostname, hostname),
		attribute.String(tracingCommonKeyIpIntranet, intranetIpStr),
		semconv.HostNameKey.String(hostname),
	}
}

// strLimit returns a portion of string `str` specified by `length` parameters, if the length
// of `str` is greater than `length`, then the `suffix` will be appended to the result string.
func strLimit(str string, length int, suffix string) string {
	if len(str) < length {
		return str
	}
	return str[0:length] + suffix
}

// DefaultRequestsCounter
// return metric.Int64Counter for WithRequests
// suggest histogramName = <client/server>_requests_code_total
func DefaultRequestsCounter(meter metric.Meter, histogramName string) (metric.Int64Counter, error) {
	return meter.Int64Counter(histogramName, metric.WithUnit("{call}"))
}

// DefaultSecondsHistogram
// return metric.Float64Histogram for WithSeconds
// suggest histogramName = <client/server>_requests_seconds_bucket
func DefaultSecondsHistogram(meter metric.Meter, histogramName string) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(histogramName, metric.WithUnit("s"))
}

// DefaultSecondsHistogramView
// need register in sdkmetric.MeterProvider
// eg:
// view := SecondsHistogramView()
// mp := sdkmetric.NewMeterProvider(sdkmetric.WithView(view))
// otel.SetMeterProvider(mp)
func DefaultSecondsHistogramView(histogramName string) metricsdk.View {
	return func(instrument metricsdk.Instrument) (metricsdk.Stream, bool) {
		if instrument.Name == histogramName {
			return metricsdk.Stream{
				Name:        instrument.Name,
				Description: instrument.Description,
				Unit:        instrument.Unit,
				Aggregation: metricsdk.AggregationExplicitBucketHistogram{
					Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
					NoMinMax:   true,
				},
				AttributeFilter: func(value attribute.KeyValue) bool {
					return true
				},
			}, true
		}
		return metricsdk.Stream{}, false
	}
}
