package main

import (
	"context"
	"time"

	"github.com/admpub/redsync"
	"github.com/admpub/redsync/redis"
	"github.com/admpub/redsync/redis/redigo"
	redigolib "github.com/gomodule/redigo/redis"
	"github.com/stvp/tempredis"
)

func main() {

	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	pool := redigo.NewPool(&redigolib.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigolib.Conn, error) {
			return redigolib.Dial("unix", server.Socket())
		},
		TestOnBorrow: func(c redigolib.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	})

	rs := redsync.New([]redis.Pool{pool})

	ctx := context.Background()
	mutex := rs.NewMutex("test-redsync")
	err = mutex.Lock(ctx)

	if err != nil {
		panic(err)
	}

	mutex.Unlock(ctx)
}
