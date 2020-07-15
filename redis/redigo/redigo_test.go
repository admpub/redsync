package redigo

import "github.com/admpub/redsync/redis"

var _ (redis.Conn) = (*Conn)(nil)

var _ (redis.Pool) = (*Pool)(nil)
