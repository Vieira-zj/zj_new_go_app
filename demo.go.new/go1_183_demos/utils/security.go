package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func CreateSignature(apiKey, apiSecret, timestamp, nonce string, params url.Values) (string, error) {
	content := apiKey + timestamp + nonce + params.Encode()
	mac := hmac.New(sha256.New, []byte(apiSecret))
	if _, err := mac.Write([]byte(content)); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func ValidateSignatureForHttpRequest(r *http.Request, inputApiKey, inputApiSecret string) bool {
	query := r.URL.Query()
	apiKey := query.Get("apiKey")
	query.Del("apiKey")
	timestamp := query.Get("timestamp")
	query.Del("timestamp")
	nonce := query.Get("nonce")
	query.Del("nonce")
	signature := query.Get("signature")
	query.Del("signature")

	if apiKey != inputApiKey {
		return false
	}

	timeInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	ti := time.Unix(timeInt, 0)
	if time.Since(ti).Seconds() > 5 {
		return false
	}

	expectedSignature, err := CreateSignature(apiKey, inputApiSecret, timestamp, nonce, query)
	if err != nil {
		log.Println("validate signature get error:", err)
		return false
	}

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
