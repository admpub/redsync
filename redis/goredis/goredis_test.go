package goredis

import "github.com/applinskinner/redsync/redis"

var _ (redis.Conn) = (*GoredisConn)(nil)

var _ (redis.Pool) = (*GoredisPool)(nil)
