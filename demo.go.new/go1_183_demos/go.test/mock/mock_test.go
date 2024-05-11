package mocktest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMockFoo01(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("pass mock case", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		Sut(mockFoo)
	})

	t.Run("invalid args", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(98)).Return(101)
		Sut(mockFoo)
	})

	t.Run("invalid calls", func(t *testing.T) {
		mockFoo := NewMockFoo(ctrl)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
		Sut(mockFoo)
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

	Sut(mockFoo)
	t.Log("foo mock test done")
}

func TestFooImpl(t *testing.T) {
	foo := NewFooImpl()
	Sut(foo)
	t.Log("foo impl done")
}

func TestMockBarGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("get int value", func(t *testing.T) {
		mockBar := NewMockBar(ctrl)
		mockBar.EXPECT().Get(gomock.Any()).Times(1).Return(9)
		s := GetString("testkey", mockBar)
		assert.Equal(t, "9", s)
	})

	t.Run("get marshal value", func(t *testing.T) {
		mockBar := NewMockBar(ctrl)
		expect := `{"name":"bar"}`
		m := make(map[string]any)
		json.Unmarshal([]byte(expect), &m)

		mockBar.EXPECT().Get(gomock.Any()).Times(1).Return(m)
		result := GetString("testkey", mockBar)
		assert.Equal(t, expect, result)
	})

	t.Run("invalid arg key", func(t *testing.T) {
		mockBar := NewMockBar(ctrl)
		mockBar.EXPECT().Get(gomock.Any()).Times(0)
		result := GetString("", mockBar)
		assert.Equal(t, "null", result)
	})
}
