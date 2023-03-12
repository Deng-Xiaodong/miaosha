package redislock

import (
	"github.com/go-redis/redis"
	"miaosha/common"
)

var RedisClient *redis.Client

func InitRedisClient(config *common.RedisConfig) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Address + ":" + config.Port,
		Password: config.Password,
	})
}
