package utils

import (
	"time"

	"github.com/allegro/bigcache/v2"
)

// GlobalCache .
var GlobalCache *bigcache.BigCache

func init() {
	var err error
	if GlobalCache, err = bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute)); err != nil {
		panic(err)
	}
}
