package demos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	retry "github.com/avast/retry-go/v3"
	"github.com/stretchr/testify/assert"
)

func TestRetryGet(t *testing.T) {
	var (
		count  = 0
		expect = "hello"
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count >= 3 {
			fmt.Fprintln(w, expect)
		} else {
			fmt.Println("mock http server error")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	var body []byte
	err := retry.Do(func() error {
		resp, err := http.Get(ts.URL)
		if err != nil {
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()

		if resp.StatusCode != 200 {
			return fmt.Errorf("non-ok http return code: %d", resp.StatusCode)
		}
		body, err = ioutil.ReadAll(resp.Body)
		return err
	}, retry.OnRetry(func(n uint, err error) {
		fmt.Printf("#retry: %d, error: %v\n", n, err)
	}), retry.Attempts(3), retry.Delay(time.Second))

	assert.NoError(t, err)
	assert.NotEmpty(t, body)
	res := strings.TrimRight(string(body), "\n")
	assert.Equal(t, expect, res)
}

//
// delay_based_on_error
//

type RetryAfterError struct {
	response http.Response
}

func (err RetryAfterError) Error() string {
	return fmt.Sprintf(
		"Request to %s fail %s (%d)",
		err.response.Request.RequestURI,
		err.response.Status,
		err.response.StatusCode,
	)
}

type SomeOtherError struct {
	err        string
	retryAfter time.Duration
}

func (err SomeOtherError) Error() string {
	return err.err
}

func TestCustomRetryFunctionBasedOnKindOfError(t *testing.T) {
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if count <= 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	var body []byte
	err := retry.Do(func() error {
		resp, err := http.Get(ts.URL)
		if err != nil {
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()

		switch resp.StatusCode {
		case http.StatusInternalServerError:
			return RetryAfterError{
				response: *resp,
			}
		case http.StatusNotFound:
			return SomeOtherError{
				err:        fmt.Sprintln("status not found"),
				retryAfter: time.Duration(2) * time.Second,
			}
		}
		body, err = ioutil.ReadAll(resp.Body)
		return err
	}, retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
		// n: 重试次数
		switch e := err.(type) {
		case RetryAfterError:
			fmt.Println("Get RetryAfterError")
			if t, err := parseRetryAfter(e.response.Request.Header.Get("Retry-After")); err == nil {
				return time.Until(t)
			}
		case SomeOtherError:
			fmt.Println("Get SomeOtherError")
			return e.retryAfter
		}
		return retry.BackOffDelay(n, err, config)
	}), retry.Attempts(4), retry.MaxDelay(time.Duration(3)*time.Second))

	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func parseRetryAfter(_ string) (time.Time, error) {
	return time.Now().Add(time.Second), nil
}
