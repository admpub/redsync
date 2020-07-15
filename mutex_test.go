package redsync

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/admpub/redsync/redis/redigo"

	"github.com/admpub/redsync/redis"
	redigoredis "github.com/gomodule/redigo/redis"
	"github.com/stvp/tempredis"
)

func TestMutex(t *testing.T) {
	pools := newMockPools(8, servers)
	mutexes := newTestMutexes(pools, "test-mutex", 8)
	orderCh := make(chan int)
	ctx := context.Background()
	for i, mutex := range mutexes {
		go func(i int, mutex *Mutex) {
			err := mutex.Lock(ctx)
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			defer mutex.Unlock(ctx)

			assertAcquired(t, pools, mutex)

			orderCh <- i
		}(i, mutex)
	}
	for range mutexes {
		<-orderCh
	}
}

func TestMutexExtend(t *testing.T) {
	pools := newMockPools(8, servers)
	mutexes := newTestMutexes(pools, "test-mutex-extend", 1)
	mutex := mutexes[0]

	ctx := context.Background()
	err := mutex.Lock(ctx)
	if err != nil {
		t.Fatalf("Expected err == nil, got %q", err)
	}
	defer mutex.Unlock(ctx)

	time.Sleep(1 * time.Second)

	expiries := getPoolExpiries(pools, mutex.name)
	ok, err := mutex.Extend(ctx)
	if err != nil {
		t.Fatalf("Expected err == nil, got %q", err)
	}
	if !ok {
		t.Fatalf("Expected ok == true, got %v", ok)
	}
	expiries2 := getPoolExpiries(pools, mutex.name)

	for i, expiry := range expiries {
		if expiry >= expiries2[i] {
			t.Fatalf("Expected expiries[%d] > expiry, got %d %d", i, expiries2[i], expiry)
		}
	}
}

func TestMutexQuorum(t *testing.T) {
	pools := newMockPools(4, servers)
	ctx := context.Background()
	for mask := 0; mask < 1<<uint(len(pools)); mask++ {
		mutexes := newTestMutexes(pools, "test-mutex-partial-"+strconv.Itoa(mask), 1)
		mutex := mutexes[0]
		mutex.tries = 1

		n := clogPools(pools, mask, mutex)

		if n >= len(pools)/2+1 {
			err := mutex.Lock(ctx)
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			assertAcquired(t, pools, mutex)
		} else {
			err := mutex.Lock(ctx)
			if err != ErrFailed {
				t.Fatalf("Expected err == %q, got %q", ErrFailed, err)
			}
		}
	}
}

func TestMutexFailure(t *testing.T) {
	var servers []*tempredis.Server
	for i := 0; i < 8; i++ {
		server, err := tempredis.Start(tempredis.Config{})
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	servers[2].Term()
	servers[6].Term()

	pools := newMockPools(8, servers)

	okayPools := []redis.Pool{}
	for i, v := range pools {
		if i == 2 || i == 6 {
			continue
		}
		okayPools = append(okayPools, v)
	}

	mutexes := newTestMutexes(pools, "test-mutex-extend", 1)
	mutex := mutexes[0]
	ctx := context.Background()

	err := mutex.Lock(ctx)
	if err != nil {
		t.Fatalf("Expected err == nil, got %q", err)
	}
	defer mutex.Unlock(ctx)

	assertAcquired(t, okayPools, mutex)
}

func TestValid(t *testing.T) {
	pools := newMockPools(8, servers)
	rs := New(pools)
	key := "test-shared-lock"

	ctx := context.Background()
	mutex1 := rs.NewMutex(key, SetExpiry(time.Hour))
	err := mutex1.Lock(ctx)
	if err != nil {
		t.Fatalf("Expected err != nil, got: %q", err)
	}
	assertAcquired(t, pools, mutex1)

	ok, err := mutex1.Valid(ctx)
	if err != nil {
		t.Fatalf("Expected err != nil, got: %q", err)
	}
	if !ok {
		t.Fatalf("Expected a valid mutex")
	}

	mutex2 := rs.NewMutex(key)
	err = mutex2.Lock(ctx)
	if err == nil {
		t.Fatalf("Expected err == nil, got: %q", err)
	}
}

func newMockPools(n int, servers []*tempredis.Server) []redis.Pool {
	pools := []redis.Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, redigo.NewPool(&redigoredis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redigoredis.Conn, error) {
					return redigoredis.Dial("unix", server.Socket())
				},
				TestOnBorrow: func(c redigoredis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}))
		}(server)
		if len(pools) == n {
			break
		}
	}
	return pools
}

func getPoolValues(pools []redis.Pool, name string) []string {
	values := []string{}
	ctx := context.Background()
	for _, pool := range pools {
		conn := pool.Get()
		value, err := conn.Get(ctx, name)
		conn.Close()
		if err != nil {
			panic(err)
		}
		values = append(values, value)
	}
	return values
}

func getPoolExpiries(pools []redis.Pool, name string) []int {
	expiries := []int{}
	ctx := context.Background()
	for _, pool := range pools {
		conn := pool.Get()
		expiry, err := conn.PTTL(ctx, name)
		conn.Close()
		if err != nil {
			panic(err)
		}
		expiries = append(expiries, int(expiry))
	}
	return expiries
}

func clogPools(pools []redis.Pool, mask int, mutex *Mutex) int {
	n := 0
	ctx := context.Background()
	for i, pool := range pools {
		if mask&(1<<uint(i)) == 0 {
			n++
			continue
		}
		conn := pool.Get()
		_, err := conn.Set(ctx, mutex.name, "foobar")
		conn.Close()
		if err != nil {
			panic(err)
		}
	}
	return n
}

func newTestMutexes(pools []redis.Pool, name string, n int) []*Mutex {
	mutexes := []*Mutex{}
	for i := 0; i < n; i++ {
		mutexes = append(mutexes, &Mutex{
			name:         name,
			expiry:       8 * time.Second,
			tries:        32,
			delayFunc:    func(tries int) time.Duration { return 500 * time.Millisecond },
			genValueFunc: genValue,
			factor:       0.01,
			quorum:       len(pools)/2 + 1,
			pools:        pools,
		})
	}
	return mutexes
}

func assertAcquired(t *testing.T, pools []redis.Pool, mutex *Mutex) {
	n := 0
	values := getPoolValues(pools, mutex.name)
	for _, value := range values {
		if value == mutex.value {
			n++
		}
	}
	if n < mutex.quorum {
		t.Fatalf("Expected n >= %d, got %d", mutex.quorum, n)
	}
}
