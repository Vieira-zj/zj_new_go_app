package jwt

import (
	"testing"
)

func TestJwtTokenUsage(t *testing.T) {
	secretKey := []byte("my-secret-key")
	token, err := CreateToken(1, "foo.bar", secretKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token:", token)

	claims, err := ParseToken(token, secretKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("claim: %+v", claims)
}
