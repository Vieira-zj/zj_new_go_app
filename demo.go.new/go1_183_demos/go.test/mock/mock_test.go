package mocktest

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestMockFoo01(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("pass mock case", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		SUT(mockFoo)
	})

	t.Run("invalid args", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(98)).Return(101)
		SUT(mockFoo)
	})

	t.Run("invalid calls", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		SUT(mockFoo)
	})

	t.Log("foo mock test done")
}

func TestMockFoo02(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFoo := NewMockFoo(ctrl)
	// Executes the anonymous functions and returns its result when Bar is invoked with 99.
	mockFoo.EXPECT().Bar(gomock.Eq(98)).DoAndReturn(func(_ int) int {
		time.Sleep(time.Second)
		return 101
	}).AnyTimes()

	SUT(mockFoo)
	t.Log("foo mock test done")
}
