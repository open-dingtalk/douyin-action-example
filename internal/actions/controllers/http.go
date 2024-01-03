package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"time"
)

type DouYinClient struct {
	httpClient *http.Client
}

func NewDouYinClient() (*DouYinClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "tcp", addr)
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 60 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}
	return &DouYinClient{
		httpClient: httpClient,
	}, nil
}

func (c *DouYinClient) Post(ctx context.Context, url string, request interface{}, response interface{}) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return c.Execute(ctx, httpRequest, response)
}

func (c *DouYinClient) Get(ctx context.Context, url string, response interface{}) error {
	httpRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	return c.Execute(ctx, httpRequest, response)

}

func (c *DouYinClient) Execute(ctx context.Context, request *http.Request, response interface{}) error {
	httpResponse, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	if httpResponse.StatusCode != http.StatusOK {
		// logger.Errorf(ctx, "status(%d) is not ok, url=%s, body(%s)", httpResponse.StatusCode, request.URL.String(), string(body))
		return errors.Errorf("status(%d) is not ok, body(%s)", httpResponse.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return nil
}
