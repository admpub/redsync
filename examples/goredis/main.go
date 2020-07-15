package main

import (
	"context"

	"github.com/admpub/redsync"
	"github.com/admpub/redsync/redis"
	"github.com/admpub/redsync/redis/goredis"
	"github.com/stvp/tempredis"
	goredislib "gopkg.in/redis.v8"
)

func main() {

	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	client := goredislib.NewClient(&goredislib.Options{
		Network: "unix",
		Addr:    server.Socket(),
	})

	pool := goredis.NewGoredisPool(client)

	rs := redsync.New([]redis.Pool{pool})

	ctx := context.Background()
	mutex := rs.NewMutex("test-redsync")
	err = mutex.Lock(ctx)

	if err != nil {
		panic(err)
	}

	mutex.Unlock(ctx)
}