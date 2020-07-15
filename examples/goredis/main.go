package main

import (
	"github.com/applinskinner/redsync"
	"github.com/applinskinner/redsync/redis"
	"github.com/applinskinner/redsync/redis/goredis"
	goredislib "github.com/go-redis/redis/v7"
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

	pool := goredis.NewGoredisPool(client)

	rs := redsync.New([]redis.Pool{pool})

	mutex := rs.NewMutex("test-redsync")
	err = mutex.Lock()

	if err != nil {
		panic(err)
	}

	mutex.Unlock()
}
