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
	"fmt"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	"github.com/opendevops-cn/codo-golang-sdk/consts"
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

func (x *NoAuthConfig) GetConfig(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", x.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request err: %w", err)
	}
	return x.client.Do(ctx, req)
}

func (x *AuthConfig) GetConfig(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", x.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request err: %w", err)
	}
	// 自动添加统一设置的 Cookie
	for _, cookie := range x.cookies {
		req.AddCookie(cookie)
	}
	return x.client.Do(ctx, req)
}
