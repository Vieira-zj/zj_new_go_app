package pkg

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

func sendHTTPRequest(c echo.Context) (*http.Response, error) {
	keyTarget := "X-Target"
	req := c.Request()
	host := req.Header.Get(keyTarget)
	req.Header.Del(keyTarget)
	url := "http://" + host + req.RequestURI

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(req.Method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header = req.Header

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
