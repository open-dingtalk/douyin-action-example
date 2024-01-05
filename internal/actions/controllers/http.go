package controllers

import (
	"context"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"strings"
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

func GetBearerToken(r *http.Request) (string, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no Authorization header")
	}

	// Split the header value by space.
	// Should be in the form of ["Bearer", "token"]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid token format, not bearer")
	}

	return parts[1], nil
}
