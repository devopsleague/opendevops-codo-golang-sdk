package xvm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
)

var (
	ErrInvalidAuth = fmt.Errorf("认证失败: 无效的用户名或密码")
	ErrQuery       = fmt.Errorf("查询失败")
)

type clientOptions struct {
	httpClient xhttp.IClient

	enableAuth bool
	auth       authConfig
}

func defaultClientOptions() clientOptions {
	httpClient, _ := xhttp.NewClient()
	return clientOptions{
		httpClient: httpClient,
	}
}

// authConfig 认证配置结构体
type authConfig struct {
	username string
	password string
}

type IMetricsClient interface {
	QueryRange(ctx context.Context, query string, opts ...IQueryRangeOption) (*QueryResult, error)
	Query(ctx context.Context, query string, queryOptions ...IQueryOption) (*QueryResult, error)
}

// MetricsClient VictoriaMetrics 客户端结构体
type MetricsClient struct {
	baseURL string
	options clientOptions
}

type QueryResult struct {
	IsPartial bool       `json:"isPartial"`
	Data      MetricData `json:"data"`
}

type IClientOption interface {
	Apply(options *clientOptions)
}

// ClientOptionFunc 客户端配置选项函数类型
type ClientOptionFunc func(*clientOptions)

func (x ClientOptionFunc) Apply(options *clientOptions) {
	x(options)
}

// WithClientOptionBasicAuth 设置 Basic Auth 认证
func WithClientOptionBasicAuth(username, password string) ClientOptionFunc {
	return func(options *clientOptions) {
		if username == "" && password == "" {
			return
		}
		options.auth = authConfig{
			username: username,
			password: password,
		}
		options.enableAuth = true
	}
}

// NewMetricsClient 创建新的 VictoriaMetrics 客户端
func NewMetricsClient(baseURL string, opts ...IClientOption) (IMetricsClient, error) {
	options := defaultClientOptions()
	// 应用配置选项
	for _, option := range opts {
		option.Apply(&options)
	}

	return &MetricsClient{
		baseURL: baseURL,
		options: options,
	}, nil
}

// addAuthHeader 添加认证头
func (c *MetricsClient) addAuthHeader(req *http.Request) {
	if c.options.enableAuth {
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.options.auth.username, c.options.auth.password)))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	}
}

type queryRangeOptions struct {
	// 开始时间
	start time.Time
	// 结束时间
	end time.Time
	// 每个时间序列最大返回的数据数量
	limit uint32
}

func defaultQueryRangeOptions() queryRangeOptions {
	return queryRangeOptions{
		start: time.Now().Add(-time.Minute * 30),
		end:   time.Now(),
		limit: 1000,
	}
}

func (x *queryRangeOptions) step() time.Duration {
	return x.end.Sub(x.start) / time.Duration(x.limit)
}

type IQueryRangeOption interface {
	Apply(options *queryRangeOptions)
}

// QueryRangeOptionFunc 查询范围选项函数类型
type QueryRangeOptionFunc func(*queryRangeOptions)

func (x QueryRangeOptionFunc) Apply(options *queryRangeOptions) {
	x(options)
}

// WithQueryRangeOptionStart 设置查询开始时间
func WithQueryRangeOptionStart(start time.Time) QueryRangeOptionFunc {
	return func(options *queryRangeOptions) {
		options.start = start
	}
}

// WithQueryRangeOptionEnd 设置查询结束时间
func WithQueryRangeOptionEnd(end time.Time) QueryRangeOptionFunc {
	return func(options *queryRangeOptions) {
		options.end = end
	}
}

// WithQueryRangeOptionLimit 设置每个时间序列最大返回的数据数量
func WithQueryRangeOptionLimit(limit uint32) QueryRangeOptionFunc {
	return func(options *queryRangeOptions) {
		if limit == 0 {
			return
		}
		options.limit = limit
	}
}

// QueryRange 查询时间范围内的指标数据
// ${query} [${start}, ${end}] interval(${step})
// query - PromQL
// start - 开始时间
// end   - 结束时间
// step  - 步长 数据点的时间间隔
func (c *MetricsClient) QueryRange(ctx context.Context, query string, opts ...IQueryRangeOption) (*QueryResult, error) {
	options := defaultQueryRangeOptions()
	for _, opt := range opts {
		opt.Apply(&options)
	}
	fullURL, err := url.JoinPath(c.baseURL, "/api/v1/query_range")
	if err != nil {
		return nil, fmt.Errorf("%w, baseURL=%s, path=%s", err, c.baseURL, "/api/v1/query_range")
	}
	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败: %v", err)
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", options.start.Unix()))
	q.Set("end", fmt.Sprintf("%d", options.end.Unix()))
	q.Set("step", fmt.Sprintf("%ds", int(options.step().Seconds())))
	// 添加 limit 参数
	q.Set("limit", fmt.Sprintf("%d", options.limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加认证头
	c.addAuthHeader(req)

	resp, err := c.options.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAuth
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 响应错误: %s, 状态码: %d", string(body), resp.StatusCode)
	}

	var result MetricResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.IsError() {
		return nil, fmt.Errorf("%w: ErrorType=%s, Error=%s", ErrQuery, result.ErrorType, result.Error)
	}

	return &QueryResult{
		IsPartial: result.IsPartial,
		Data:      result.Data,
	}, nil
}

type queryOptions struct {
	timestamp      time.Time
	queryTimestamp bool

	// 每个时间序列最大返回的数据数量
	limit uint32
}

func defaultQueryOptions() queryOptions {
	return queryOptions{
		timestamp:      time.Unix(0, 0),
		queryTimestamp: false,
		limit:          1000,
	}
}

type IQueryOption interface {
	Apply(options *queryOptions)
}

// QueryOptionFunc 查询选项函数类型
type QueryOptionFunc func(*queryOptions)

func (x QueryOptionFunc) Apply(options *queryOptions) {
	x(options)
}

// WithQueryOptionTimestamp 设置查询时间戳
func WithQueryOptionTimestamp(timestamp time.Time) QueryOptionFunc {
	return func(options *queryOptions) {
		options.timestamp = timestamp
		options.queryTimestamp = true
	}
}

// WithQueryOptionLimit 设置每个时间序列最大返回的数据数量
func WithQueryOptionLimit(limit uint32) QueryOptionFunc {
	return func(options *queryOptions) {
		options.limit = limit
	}
}

// Query 执行即时查询
func (c *MetricsClient) Query(ctx context.Context, query string, queryOptions ...IQueryOption) (*QueryResult, error) {
	options := defaultQueryOptions()
	for _, opt := range queryOptions {
		opt.Apply(&options)
	}

	fullURL, err := url.JoinPath(c.baseURL, "/api/v1/query")
	if err != nil {
		return nil, fmt.Errorf("%w, baseURL=%s, path=%s", err, c.baseURL, "/api/v1/query")
	}
	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	if options.queryTimestamp {
		q.Set("time", fmt.Sprintf("%d", options.timestamp.Unix()))
	}
	// 添加 limit 参数
	q.Set("limit", fmt.Sprintf("%d", options.limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加认证头
	c.addAuthHeader(req)

	resp, err := c.options.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAuth
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 响应错误: %s, 状态码: %d", string(body), resp.StatusCode)
	}

	var result MetricResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.IsError() {
		return nil, fmt.Errorf("%w: ErrorType=%s, Error=%s", ErrQuery, result.ErrorType, result.Error)
	}

	return &QueryResult{
		IsPartial: result.IsPartial,
		Data:      result.Data,
	}, nil
}
