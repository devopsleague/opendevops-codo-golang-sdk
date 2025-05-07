/*
@Time : 2025/5/7 12:27
@Author : dongdongliu
@File : k2
@Version: 1.0
@Software: GoLand
*/
package k2

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"github.com/opendevops-cn/codo-golang-sdk/consts"
	"io"
	"net/http"
)

type NoAuthConfig struct {
	url    string
	client xhttp.IClient
}

type AuthConfig struct {
	url     string
	authKey string
	client  xhttp.IClient
	cookies []*http.Cookie
}

type Response struct {
	Code   uint32            `json:"code"`
	Msg    string            `json:"msg"`
	Reason string            `json:"reason"`
	Data   map[string]string `json:"data"`
}

func NewNoAuthConfig(url string, client xhttp.IClient) *NoAuthConfig {
	return &NoAuthConfig{
		url:    url,
		client: client,
	}
}

func NewAuthConfig(url, authKey string, client xhttp.IClient) *AuthConfig {
	Cookie := &http.Cookie{
		Name:  consts.CODOAPIGatewayAuthKeyHeader,
		Value: authKey,
	}
	return &AuthConfig{
		url:     url,
		client:  client,
		cookies: []*http.Cookie{Cookie},
	}
}

func getConfig(ctx context.Context, client xhttp.IClient, url string, cookies []*http.Cookie) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request err: %w", err)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("do request err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body err: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal body err: %w", err)
	}
	if apiResp.Code != consts.CodoAPISuccessCode {
		return nil, fmt.Errorf("response code %d", apiResp.Code)
	}
	return apiResp.Data, nil
}

func (x *NoAuthConfig) GetConfig(ctx context.Context) (map[string]string, error) {
	return getConfig(ctx, x.client, x.url, nil)
}

func (x *AuthConfig) GetConfig(ctx context.Context) (map[string]string, error) {
	return getConfig(ctx, x.client, x.url, x.cookies)
}
