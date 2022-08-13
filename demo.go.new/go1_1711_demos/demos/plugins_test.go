package demos

import (
	"context"
	"fmt"
	"testing"
	"time"
)

//
// Demo: process with register middlewares / plugins
//

type handler func(context.Context)

type middleware func(context.Context, handler)

type processor struct {
	handler     handler
	middlewares []middleware
}

func newProcessor(h handler) *processor {
	return &processor{
		handler:     h,
		middlewares: make([]middleware, 0, 4),
	}
}

func (p *processor) registerMiddleware(mware middleware) {
	p.middlewares = append(p.middlewares, mware)
}

func (p *processor) exec(ctx context.Context) {
	if len(p.middlewares) == 0 {
		p.handler(ctx)
		return
	}
	mware := p.popMiddleware()
	mware(ctx, p.exec)
}

func (p *processor) popMiddleware() middleware {
	mware := p.middlewares[0]
	p.middlewares = p.middlewares[1:]
	return mware
}

func middlewareOne(ctx context.Context, next handler) {
	fmt.Println("middlewareOne before")
	start := time.Now()
	next(ctx)
	duration := time.Since(start).Milliseconds()
	fmt.Printf("middlewareOne after, duration: %d\n", duration)
}

func middlewareTwo(ctx context.Context, next handler) {
	fmt.Println("middlewareTwo before, and sleep")
	time.Sleep(200 * time.Millisecond)
	next(ctx)
	fmt.Println("middlewareTwo after")
}

func middlewareThree(ctx context.Context, next handler) {
	fmt.Println("middlewareThree before")
	next(ctx)
	fmt.Println("middlewareThree after")
}

func TestProcessorWithMiddlewares(t *testing.T) {
	h := func(ctx context.Context) {
		time.Sleep(time.Second)
		fmt.Println("handler exec")
	}
	p := newProcessor(h)

	p.registerMiddleware(middlewareOne)
	p.registerMiddleware(middlewareTwo)
	p.registerMiddleware(middlewareThree)
	p.exec(context.TODO())
	fmt.Println("done")
}
