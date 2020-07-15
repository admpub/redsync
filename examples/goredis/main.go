package main

import (
	"context"

	"github.com/admpub/redsync"
	"github.com/admpub/redsync/redis"
	"github.com/admpub/redsync/redis/goredis"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/stvp/tempredis"
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

	pool := goredis.NewPool(client)

	rs := redsync.New([]redis.Pool{pool})

	ctx := context.Background()
	mutex := rs.NewMutex("test-redsync")
	err = mutex.Lock(ctx)

	if err != nil {
		panic(err)
	}

	mutex.Unlock(ctx)
}
