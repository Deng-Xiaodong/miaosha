package redislock

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"
)

type Keys struct {
	Key   string
	Value string
}

func SetKeys(keys ...Keys) {
	for _, key := range keys {
		RedisClient.SetNX(key.Key, key.Value, 0)
	}
}

func getOne() bool {
	//uuid作为锁标记，锁只能由加锁的人去解
	sample := uuid.NewString()
	if RedisClient.SetNX("LOCK", sample, 10*time.Second).Val() {

		inventory, _ := RedisClient.Get("inventory").Int()
		if inventory > 0 {
			RedisClient.Set("inventory", strconv.Itoa(inventory-1), 0)
		}

		var luaScript = redis.NewScript(`
		if redis.call("get",KEYS[1]) == ARGV[1] 
		then
			redis.call("del",KEYS[1]) 
			return true
		else 
			return false 
		end
	`)
		//执行脚本
		r, err := luaScript.Run(RedisClient, []string{"LOCK"}, []string{sample}).Bool()
		if err != nil {
			log.Fatalln(err)
		}
		return r
	}
	return false
}
func GetOne() bool {
	//每个请求有两秒时间抢锁
	c := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-c:
			return false
		default:
			if getOne() {
				return true
			}
		}
	}
}
