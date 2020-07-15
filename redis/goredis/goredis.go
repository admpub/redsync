package goredis

import (
	"context"
	"strings"
	"time"

	redsyncredis "github.com/admpub/redsync/redis"
	"github.com/go-redis/redis/v8"
)

type Pool struct {
	delegate *redis.Client
}

func (self *Pool) Get() redsyncredis.Conn {
	return &Conn{self.delegate}
}

func NewPool(delegate *redis.Client) *Pool {
	return &Pool{delegate}
}

type Conn struct {
	delegate *redis.Client
}

func (self *Conn) Get(ctx context.Context, name string) (string, error) {
	value, err := self.delegate.Get(ctx, name).Result()
	err = noErrNil(err)
	return value, err
}

func (self *Conn) Set(ctx context.Context, name string, value string) (bool, error) {
	reply, err := self.delegate.Set(ctx, name, value, 0).Result()
	return err == nil && reply == "OK", err
}

func (self *Conn) SetNX(ctx context.Context, name string, value string, expiry time.Duration) (bool, error) {
	return self.delegate.SetNX(ctx, name, value, expiry).Result()
}

func (self *Conn) PTTL(ctx context.Context, name string) (time.Duration, error) {
	return self.delegate.PTTL(ctx, name).Result()
}

func (self *Conn) Eval(ctx context.Context, script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	var keys []string
	var args []interface{}

	if script.KeyCount > 0 {

		keys = []string{}

		for i := 0; i < script.KeyCount; i++ {
			keys = append(keys, keysAndArgs[i].(string))
		}

		args = keysAndArgs[script.KeyCount:]

	} else {
		keys = []string{}
		args = keysAndArgs
	}

	v, err := self.delegate.EvalSha(ctx, script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = self.delegate.Eval(ctx, script.Src, keys, args...).Result()
	}
	err = noErrNil(err)
	return v, err
}

func (self *Conn) Close() error {
	// Not needed for this library
	return nil
}

func noErrNil(err error) error {
	if err != nil && err.Error() == "redis: nil" {
		return nil
	}
	return err
}
