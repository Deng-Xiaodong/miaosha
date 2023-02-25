package redislock

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"miaosha/common"
	"strconv"
	"time"
)

var Client *redis.Client

const INV = 10

func InitRedisClient(config *common.Config) {
	Client = redis.NewClient(&redis.Options{
		Addr:     config.RedCfg.Address + ":" + config.RedCfg.Port,
		Password: config.RedCfg.Password,
	})
}

func getOne() (n int, b bool) {
	sample := uuid.NewString()
	if Client.SetNX("LOCK", sample, 10*time.Second).Val() {
		var v int
		s := Client.Get("inventory").Val()
		if s == "" {
			v = INV
			Client.Set("inventory", strconv.Itoa(v), 0)
		} else {
			v, _ = strconv.Atoi(s)
		}
		if v > 0 {
			Client.Set("inventory", strconv.Itoa(v-1), 0)
		}
		//编写脚本 - 检查数值，是否够用，够用再减，否则返回减掉后的结果
		var luaScript = redis.NewScript(`
		if redis.call("get",KEYS[1]) == ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end
	`)
		//执行脚本
		luaScript.Run(Client, []string{"LOCK"}, []string{sample})
		return v, true
	}
	return -1, false
}
func GetLock() int {
	c := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-c:
			return -1
		default:
			if n, ok := getOne(); ok {
				return n
			}
		}
	}
}
