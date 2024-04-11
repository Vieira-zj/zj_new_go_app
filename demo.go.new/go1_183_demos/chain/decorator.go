package chain

import (
	"context"
	"log"
)

type handleFunc func(ctx context.Context, param map[string]any) error

func Decorate(fn handleFunc) handleFunc {
	return func(ctx context.Context, param map[string]any) error {
		log.Println("pre-process...")
		err := fn(ctx, param)
		log.Println("post-process...")
		return err
	}
}

// Chain

type DecorateFunc func(ctx context.Context, param map[string]any, fn handleFunc) error

func DecoratorsChain(decorators []DecorateFunc) DecorateFunc {
	return func(ctx context.Context, param map[string]any, fn handleFunc) error {
		return decorators[0](ctx, param, getChainHandler(decorators, 0, fn))
	}
}

func getChainHandler(decorators []DecorateFunc, curr int, fn handleFunc) handleFunc {
	if curr == len(decorators)-1 {
		return fn
	}

	return func(ctx context.Context, param map[string]any) error {
		return decorators[curr+1](ctx, param, getChainHandler(decorators, curr+1, fn))
	}
}
