package redislock

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"
)

type DisLock struct {
	RedisClient *redis.Client
}

func NewDisLock(redc *redis.Client) *DisLock {
	return &DisLock{redc}
}

type Keys struct {
	Key   string
	Value string
}

func SetKeys(redc *redis.Client, keys ...Keys) {
	for _, key := range keys {
		redc.SetNX(key.Key, key.Value, 0)
	}
}

func (dl *DisLock) getOne() bool {
	//uuid作为锁标记，锁只能由加锁的人去解
	sample := uuid.NewString()
	if dl.RedisClient.SetNX("LOCK", sample, 10*time.Second).Val() {

		inventory, _ := dl.RedisClient.Get("inventory").Int()
		if inventory > 0 {
			dl.RedisClient.Set("inventory", strconv.Itoa(inventory-1), 0)
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
		r, err := luaScript.Run(dl.RedisClient, []string{"LOCK"}, []string{sample}).Bool()
		if err != nil {
			log.Fatalln(err)
		}
		return r
	}
	return false
}
func (dl *DisLock) GetOne() bool {
	//每个请求有两秒时间抢锁
	c := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-c:
			return false
		default:
			if dl.getOne() {
				return true
			}
		}
	}
}
