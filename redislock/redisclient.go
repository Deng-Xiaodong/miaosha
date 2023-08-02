package redislock

import (
	"github.com/go-redis/redis"
	"miaosha/common"
)

//var RedisClient *redis.Client

func InitRedisClient(config *common.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Address + ":" + config.Port,
		Password: config.Password,
	})
}
