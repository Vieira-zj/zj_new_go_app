package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
)

// SampleHandler01 test handler.
func SampleHandler01(c echo.Context) error {
	base := c.QueryParam("base")
	if len(base) == 0 {
		base = "10"
	}

	num, err := strconv.Atoi(base)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	outc, errc := asyncService(c, num)
	select {
	case res := <-outc:
		return c.String(http.StatusOK, strconv.Itoa(res))
	case err := <-errc:
		// when close(errc), err is nil
		if err != nil {
			return c.String(http.StatusOK, err.Error())
		}
	}
	return nil
}

func asyncService(c echo.Context, base int) (<-chan int, <-chan error) {
	outc := make(chan int)
	errc := make(chan error, 1)

	go func() {
		defer close(outc)
		defer close(errc)

		if base <= 0 {
			errc <- fmt.Errorf("invalid input base")
			return
		}
		if base > 10 {
			base = 10
		}

		for i := 0; i < base; i++ {
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
			select {
			case <-c.Request().Context().Done():
				c.Logger().Info("request context cancel")
				return
			default:
			}
		}
		outc <- base
	}()

	return outc, errc
}
