package goredis

import "github.com/admpub/redsync/redis"

var _ (redis.Conn) = (*GoredisConn)(nil)

var _ (redis.Pool) = (*GoredisPool)(nil)