package sdk

import (
	"strconv"

	"github.com/go-redis/redis"
)

type Param struct {
	Address  string
	Password string
	DB       string
}

func (c *Param) New(impl *Redis) (*Redis, error) {
	if c.DB == "" {
		c.DB = "0"
	}
	dbInt, err := strconv.Atoi(c.DB)
	if err != nil {
		return impl, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Password: c.Password,
		DB:       dbInt,
	})

	_, err = client.Ping().Result()
	if err != nil {
		return impl, err
	}

	impl.client = client
	return impl, nil
}
