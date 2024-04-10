package algorithm

import (
	"context"
	"errors"
	"log"
	"testing"
)

func TestDecorate(t *testing.T) {
	fn := func(ctx context.Context, param map[string]any) error {
		if len(param) == 0 {
			return errors.New("no param input")
		}
		for k, v := range param {
			log.Printf("get: key=%s, value=%v", k, v)
		}
		return nil
	}

	dfn := Decorate(fn)
	param := map[string]any{"one": 1, "two": 2}
	if err := dfn(context.TODO(), param); err != nil {
		t.Fatal(err)
	}

	t.Log("decorate done")
}

func TestChainDecorators(t *testing.T) {
	dfn1 := func(ctx context.Context, param map[string]any, fn handleFunc) error {
		log.Println("[decorator1] pre-process...")
		err := fn(ctx, param)
		log.Println("[decorator1] post-process...")
		return err
	}
	dfn2 := func(ctx context.Context, param map[string]any, fn handleFunc) error {
		log.Println("[decorator2] pre-process...")
		err := fn(ctx, param)
		log.Println("[decorator2] post-process...")
		return err
	}

	chain := ChainDecorators([]DecorateFunc{dfn1, dfn2})

	// invoke
	fn := func(ctx context.Context, param map[string]any) error {
		if len(param) == 0 {
			return errors.New("no param input")
		}
		for k, v := range param {
			log.Printf("get: key=%s, value=%v", k, v)
		}
		return nil
	}
	if err := chain(context.TODO(), map[string]any{"one": 1, "two": 2}, fn); err != nil {
		t.Fatal(err)
	}

	t.Log("chain decorator done")
}
