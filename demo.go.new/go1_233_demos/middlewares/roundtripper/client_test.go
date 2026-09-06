package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHttpClientWithRoundTripper(t *testing.T) {
	url := "http://zj.test.com:8081/health"
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, url, strings.NewReader("{}"))
	require.NoError(t, err, "new http request failed")

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
	require.NoError(t, err, "http request failed")
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "read response body failed")
	slog.Info("response", slog.Int("status", resp.StatusCode), slog.String("body", string(b)))
}
