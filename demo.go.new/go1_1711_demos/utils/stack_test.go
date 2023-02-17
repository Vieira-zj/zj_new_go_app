package utils

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSyncStackAddAndPop(t *testing.T) {
	stack := NewSyncStack(10)

	for i := 0; i < 3; i++ {
		err := stack.Add(context.Background(), i)
		assert.NoError(t, err)
	}
	t.Log("elements:", stack.String())

	ele, err := stack.Pop(context.Background())
	assert.NoError(t, err)
	t.Log("pop:", ele)
	t.Log("elements:", stack.String())
}

func TestSyncStackAddWithCancel(t *testing.T) {
	stack := NewSyncStack(10)
	defer func() {
		t.Log("close")
		stack.Close()
		time.Sleep(100 * time.Millisecond)
	}()

	for i := 0; i < 3; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			_, err := stack.Pop(ctx)
			assert.Equal(t, context.DeadlineExceeded, errors.Cause(err))
			t.Log("pop cancelled")
		}()
	}

	time.Sleep(200 * time.Millisecond)
	stack.Add(context.Background(), 1)
	assert.Equal(t, 1, stack.Size())
	t.Log("add element")
	time.Sleep(100 * time.Millisecond)
}
