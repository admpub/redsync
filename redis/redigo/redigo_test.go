package redigo

import "github.com/admpub/redsync/redis"

var _ (redis.Conn) = (*RedigoConn)(nil)

var _ (redis.Pool) = (*RedigoPool)(nil)
