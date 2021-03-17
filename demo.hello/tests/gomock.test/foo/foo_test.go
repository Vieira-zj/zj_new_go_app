package foo

import (
	"testing"
	"time"

	"demo.hello/tests/gomock.test/foo/mock_foo"
	"github.com/golang/mock/gomock"
)

func TestFoo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFoo := mock_foo.NewMockFoo(ctrl)
	mockFoo.EXPECT().Bar(99).Return(100)
	PrintFoo(mockFoo, 99)

	mockFoo.EXPECT().Bar(gomock.Eq(99)).Return(101)
	PrintFoo(mockFoo, 99)

	mockFoo.EXPECT().Bar(gomock.Any()).DoAndReturn(func(_ int) int {
		time.Sleep(time.Second)
		return 102
	}).AnyTimes()
	PrintFoo(mockFoo, 98)
}
