package chain

import (
	"context"
	"log"
)

type Interceptor interface {
	apply(*InterceptorsChain) error
}

type InterceptorsChain struct {
	index        int
	ctx          context.Context
	interceptors []Interceptor
	errors       []error
}

func NewInterceptorsChain(index int, interceptors []Interceptor) *InterceptorsChain {
	return &InterceptorsChain{
		index:        index,
		ctx:          context.Background(),
		interceptors: interceptors,
		errors:       make([]error, 0),
	}
}

func (chain *InterceptorsChain) Next() {
	if chain.index >= len(chain.interceptors) {
		log.Println("end of chain")
		return
	}

	interceptor := chain.interceptors[chain.index]
	chain.index += 1
	if err := interceptor.apply(chain); err != nil {
		chain.errors = append(chain.errors, err)
	}
}
