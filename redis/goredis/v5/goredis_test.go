package goredis5

import "github.com/admpub/redsync/v4/redis"

var _ (redis.Conn) = (*Conn)(nil)

var _ (redis.Pool) = (*Pool)(nil)
