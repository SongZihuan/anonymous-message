package reqrate

import (
	"context"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis() error {
	if rdb != nil {
		return nil
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     flagparser.RedisAddress,
		Password: flagparser.RedisPassword, // no password set
		DB:       flagparser.RedisDB,       // use default DB
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	return nil
}

func CloseRedis() {
	if rdb == nil {
		return
	}

	_ = rdb.Close()
	rdb = nil
}
