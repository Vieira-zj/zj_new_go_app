package jwt_test

import (
	"fmt"
	"testing"
	"time"

	"demo.apps/middlewares/jwt"
)

func TestJwtTokenUsage(t *testing.T) {
	token, err := jwt.CreateToken("1010", "foo.bar@test.com", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token:", token)

	claims, err := jwt.ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("claim: %+v", *claims)

	expireAt, err := claims.GetExpirationTime()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("is expired:", expireAt.Before(time.Now()))
}

func TestJwtTokenExpired(t *testing.T) {
	token, err := jwt.CreateToken("1010", "foo.bar@test.com", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token:", token)

	time.Sleep(3 * time.Second)
	_, err = jwt.ParseToken(token)
	if err == nil {
		t.Fatal(fmt.Errorf("token should be expired"))
	}
	t.Log(err.Error())
}
