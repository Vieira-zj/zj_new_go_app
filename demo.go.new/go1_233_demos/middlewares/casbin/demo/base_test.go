package casbin

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
)

const (
	ActRead  = "read"
	ActWrite = "write"
)

// Doc: https://casbin.org/docs/get-started

func TestCasbinDemo(t *testing.T) {
	enforcer, err := casbin.NewEnforcer("./model.conf", "./policy.csv")
	assert.NoError(t, err)

	for _, args := range [][]string{
		{"alice", "data1", ActRead},
		{"alice", "data1", ActWrite},
		{"bob", "data1", ActRead},
		{"bob", "data1", ActWrite},
	} {
		sub, obj, act := args[0], args[1], args[2]
		ok, err := enforcer.Enforce(sub, obj, act)
		assert.NoError(t, err)
		t.Log(sub, obj, act, ok)
	}
}
