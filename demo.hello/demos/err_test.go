package demos

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

//
// Error check
//

func TestErr01(t *testing.T) {
	errDivideByZero := errors.New("divide by zero")

	divide := func(a, b int) (int, error) {
		if b == 0 {
			return 0, errDivideByZero
		}
		return a / b, nil
	}

	a, b := 10, 0
	result, err := divide(a, b)
	if err != nil {
		switch {
		case errors.Is(err, errDivideByZero):
			t.Fatal("divide by zero error")
		default:
			t.Fatalf("unexpected division error: %s\n", err)
		}
	}
	fmt.Printf("%d / %d = %d\n", a, b, result)
}

//
// Custom error
//

type divisionError struct {
	a   int
	b   int
	msg string
}

func (err divisionError) Error() string {
	return err.msg
}

func TestErr02(t *testing.T) {
	divide := func(a, b int) (int, error) {
		if b == 0 {
			msg := fmt.Sprintf("cannot divide '%d' by zero", a)
			return 0, divisionError{
				a:   a,
				b:   b,
				msg: msg,
			}
		}
		return a / b, nil
	}

	a, b := 10, 0
	result, err := divide(a, b)
	if err != nil {
		var divErr divisionError
		switch {
		// why &divErr ?
		case errors.As(err, &divErr):
			t.Fatalf("%d / %d is not mathematically valid: %s\n", divErr.a, divErr.b, divErr.msg)
		default:
			t.Fatalf("unexpected division error: %s\n", err)
		}
	}
	fmt.Printf("%d / %d = %d\n", a, b, result)
}

//
// Wrap error
//
// 每次从一个函数中收到错误并想继续将其返回到函数链中时，至少用函数的名称来包裹错误。
//

type TestErrUser struct {
	name string
	age  int
}

func FindUser(username string) (*TestErrUser, error) {
	if strings.Contains(strings.ToLower(username), "mock") {
		notFoundErr := fmt.Errorf("user [%s] not found", username)
		return nil, fmt.Errorf("FindUser: failed executing db query: %w", notFoundErr)
	}

	rand.Seed(time.Now().UnixNano())
	return &TestErrUser{
		name: username,
		age:  rand.Intn(45),
	}, nil
}

func SetUserAge(u *TestErrUser, age int) error {
	if age > 45 {
		invalidErr := fmt.Errorf("invalid age (exceed 45): %d", age)
		return fmt.Errorf("SetUserAge: failed executing db update: %w", invalidErr)
	}
	return nil
}

func FindAndSetUserAge(username string, age int) error {
	user, err := FindUser(username)
	if err != nil {
		return fmt.Errorf("FindAndSetUserAge: %w", err)
	}

	if err := SetUserAge(user, age); err != nil {
		return fmt.Errorf("FindAndSetUserAge: %w", err)
	}
	return nil
}

func TestErr03(t *testing.T) {
	if err := FindAndSetUserAge("mock-foo", 40); err != nil {
		t.Fatalf("failed finding or updating user: %s", err)
	}
	fmt.Println("successfully updated user's age")
}

//
// Unwrap
//

func TestErr04(t *testing.T) {
	mockErr := errors.New("mock error")

	getWrapErr := func() error {
		return fmt.Errorf("wrap: %w", mockErr)
	}

	// wrap error check
	err := getWrapErr()
	if err != nil {
		if errors.Is(err, mockErr) {
			fmt.Println("true:", err)
		}
	}

	// unwrap error
	originErr := errors.Unwrap(err)
	fmt.Println("origin:", originErr)
}
