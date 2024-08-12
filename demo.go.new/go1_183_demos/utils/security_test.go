package utils_test

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

const (
	apiKey    = "your_api_key"
	apiSecret = "your_api_secret"
)

func TestValidateSignatureForHttpRequest(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "random_nonce"

	query := url.Values{}
	query.Set("param1", "value1")
	query.Set("param2", "value2")

	signature, err := utils.CreateSignature(apiKey, apiSecret, timestamp, nonce, query)
	assert.NoError(t, err)

	req, err := http.NewRequest("GET", "http://example.com/api", nil)
	assert.NoError(t, err)

	query.Set("apiKey", apiKey)
	query.Set("timestamp", timestamp)
	query.Set("nonce", nonce)
	query.Set("signature", signature)
	req.URL.RawQuery = query.Encode()

	results := utils.ValidateSignatureForHttpRequest(req, apiKey, apiSecret)
	t.Log("validate signature results:", results)
	assert.True(t, results)
}
