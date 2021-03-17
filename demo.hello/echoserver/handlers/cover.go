package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

// CoverHandler test for code coverage.
func CoverHandler(c echo.Context) error {
	cond1 := c.QueryParam("cond1")
	cond2 := c.QueryParam("cond2")
	return c.String(http.StatusOK, getCondition(c, cond1, cond2))
}

func getCondition(c echo.Context, cond1, cond2 string) string {
	ret := ""
	if strings.ToLower(cond1) == "true" {
		c.Logger().Info("run condition A")
		ret = "A"
	} else {
		c.Logger().Info("run condition B")
		ret = "B"
	}

	if strings.ToLower(cond2) == "true" {
		c.Logger().Info("run condition C")
		ret += "C"
	} else {
		c.Logger().Info("run condition D")
		ret += "D"
	}
	return ret
}
