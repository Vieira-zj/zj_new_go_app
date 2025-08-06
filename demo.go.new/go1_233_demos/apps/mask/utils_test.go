package mask

import "testing"

func TestSensitiveUtils(t *testing.T) {
	user := User{
		Name:     "Foo",
		Password: "abcd1234",
		Phone:    "1234567890",
		Email:    "foo.bar@google.com",
	}

	t.Run("output sensitive", func(t *testing.T) {
		Output(user)
	})

	t.Run("make sensitive", func(t *testing.T) {
		u := MakeSensitive(user)
		t.Logf("user: %+v", u)
	})
}
