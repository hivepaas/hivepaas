package redishelper

import "github.com/redis/go-redis/v9"

type Cmdable interface {
	redis.Cmdable
}
