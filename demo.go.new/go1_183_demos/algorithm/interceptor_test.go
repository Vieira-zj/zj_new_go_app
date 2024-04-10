package algorithm

import (
	"testing"
)

func TestMyInterceptorChain(t *testing.T) {
	interceptors := []Interceptor{
		&MyInterceptorOne{
			name: "1st-interceptor",
		},
		&MyInterceptorErr{
			name:    "err-interceptor",
			isError: false,
		},
		&MyInterceptorTwo{
			name: "2nd-interceptor",
		},
	}

	chain := NewMyInterceptorChain(0, interceptors)
	chain.Proceed()

	for _, err := range chain.errors {
		t.Log("chain error:", err)
	}
	t.Log("chain done")
}
