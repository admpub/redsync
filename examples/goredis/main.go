package main

import (
	"github.com/admpub/redsync/v4"
	"github.com/admpub/redsync/v4/redis/goredis"
	goredislib "github.com/go-redis/redis"
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

	rs := redsync.New(pool)

	ctx := context.Background()
	mutex := rs.NewMutex("test-redsync")

	if err = mutex.Lock(); err != nil {
		panic(err)
	}

	if _, err = mutex.Unlock(); err != nil {
		panic(err)
	}
}
