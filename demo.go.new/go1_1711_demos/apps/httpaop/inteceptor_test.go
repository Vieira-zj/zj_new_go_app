package httpaop

import (
	"context"
	"go1_1711_demo/utils"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWithMyHttpTransport01(t *testing.T) {
	// http.DefaultTransport warpped with MyHttpTransport
	WithMyHttpTransport(nil, nil)

	url := "http://localhost:8082/ping"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	t.Logf("status code: %d", resp.StatusCode)
	t.Logf("resp body: %s", b)
}

func TestWithMyHttpTransport02(t *testing.T) {
	// http.Client Transport wrapped with MyHttpTransport
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	WithMyHttpTransport(client, nil)

	url := "http://localhost:8082/ping"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	t.Logf("status code: %d", resp.StatusCode)
	t.Logf("resp body: %s", b)
}

func TestWithMyHttpTransport03(t *testing.T) {
	// custom client Transport wrapped with MyHttpTransport
	httpClient := utils.NewDefaultHttpRequester()
	rawClient := httpClient.GetClient()
	WithMyHttpTransport(rawClient, rawClient.Transport)

	url := "http://localhost:8082/ping"
	header := map[string]string{
		"Content-Type": "application/json",
		"X-Tag":        "http-interceptor",
	}
	resp, body, err := httpClient.Get(context.Background(), url, header)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status code: %d", resp.StatusCode)
	t.Logf("resp body: %s", body)
}
