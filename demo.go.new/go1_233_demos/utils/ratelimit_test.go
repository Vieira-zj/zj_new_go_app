package utils_test

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/utils"
)

func TestDebounce(t *testing.T) {
	getFn := func() func() {
		iterator := 0
		return func() {
			iterator += 1
			fmt.Printf("process run at %d\n", iterator)
		}
	}

	synctest.Test(t, func(st *testing.T) {
		eventCh := make(chan struct{})
		// cannot use t.Context() since it is canceled outside the bubble, and get error:
		// fatal error: close of synctest channel from outside bubble
		debounce, err := utils.NewDebounce(context.TODO(), time.Second, eventCh, getFn())
		assert.NoError(st, err)
		defer debounce.Stop()

		eventCh <- struct{}{}

		for i := range 5 {
			sleep := (i + 1) * 300
			if sleep > 1200 {
				sleep = 300
			}
			st.Logf("sleep %d and send event", sleep)
			synctest.Sleep(time.Duration(sleep) * time.Millisecond)
			eventCh <- struct{}{}
		}
		synctest.Sleep(2 * time.Second)
		st.Log("finished")
	})
}
