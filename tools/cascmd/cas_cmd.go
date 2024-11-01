package cascmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/ccheers/xpkg/sync/try_lock"
)

func NewCasCmd(client *redis.Client) try_lock.CASCommand {
	return &RedisImpl{client: client}
}

type RedisImpl struct {
	client *redis.Client
}

// CAS
// 如果 src 不等于 key 对应的值，则 false
// 如果 src 等于 key 对应的值，则赋予新值并返回 true
func (x *RedisImpl) CAS(key, src, dst string) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	// WATCH命令开始监视一个键
	err := x.client.Watch(ctx, func(tx *redis.Tx) error {
		// 在事务内部获取当前键的值
		currentValue, err := tx.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			// 获取出错，可能是键不存在或者其他错误
			return err
		}

		// 如果当前值等于期望的旧值，则进行设置新值
		if currentValue == src {
			// MULTI命令开始一个事务
			_, err := tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				// 如果目的是删除 则直接删除
				if dst == "" {
					return x.client.Del(ctx, key).Err()
				}
				// 设置新值
				return pipe.Set(ctx, key, dst, time.Minute).Err()
			})
			return err
		}
		// 如果当前值不是期望的旧值，则不作任何操作
		return fmt.Errorf("current value is not equal to src")
	}, key) // WATCH监视的键

	return err == nil
}
