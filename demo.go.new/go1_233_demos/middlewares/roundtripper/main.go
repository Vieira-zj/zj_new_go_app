package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func main() {
	url := "http://zj.test.com"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		panic(fmt.Errorf("new http request failed: %v", err))
	}

	token := func() string {
		return "mocked_token"
	}
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: Chain(
			http.DefaultTransport, Logging(), Retry(3), Auth(token),
		),
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("http request failed: %v", err))
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	slog.Info("response", slog.Int("status", resp.StatusCode), slog.String("body", string(b)))
}

// Middlewares

func Auth(token func() string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(req *http.Request) (*http.Response, error) {
			r := req.Clone(req.Context())
			r.Header.Set("Authorization", "Bearer "+token())
			return next.RoundTrip(r)
		})
	}
}

func Logging() Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := next.RoundTrip(req)
			code := 0
			if resp != nil {
				code = resp.StatusCode
			}
			slog.InfoContext(req.Context(), "http client",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", code),
				slog.Int64("cost", int64(time.Since(start))),
				slog.String("error", err.Error()),
			)
			return resp, err
		})
	}
}

func Retry(maxRetry int) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return TransportFunc(func(req *http.Request) (resp *http.Response, err error) {
			for i := 0; i <= maxRetry; i++ {
				r, e := cloneForRetry(req)
				if e != nil {
					return nil, e
				}
				resp, err = next.RoundTrip(r)
				if !retryable(resp, err) || i == maxRetry {
					return resp, err
				}
				closeBody(resp)
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			}
			return
		})
	}
}

func cloneForRetry(req *http.Request) (*http.Request, error) {
	r := req.Clone(req.Context())
	if req.Body == nil || req.Body == http.NoBody {
		return r, nil
	}
	if req.GetBody == nil {
		return nil, fmt.Errorf("body not replayable")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	r.Body = body
	return r, nil
}

func closeBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512<<10))
	_ = resp.Body.Close()
}

func retryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
}

// Helper

type TransportFunc func(*http.Request) (*http.Response, error)

func (f TransportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type Middleware func(http.RoundTripper) http.RoundTripper

func Chain(rt http.RoundTripper, ms ...Middleware) http.RoundTripper {
	for i := len(ms) - 1; i >= 0; i-- {
		rt = ms[i](rt)
	}
	return rt
}
