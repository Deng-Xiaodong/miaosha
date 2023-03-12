package redislock

import (
	"github.com/go-redis/redis"
	"log"
	"time"
)

type Limit struct {
	key   string
	rate  int
	burst int
}

func NewLimit(key string, rate, burst int) *Limit {
	return &Limit{
		key:   key,
		rate:  rate,
		burst: burst,
	}
}

const luaScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local tokens = tonumber(redis.call('get', key) or 0)
local last = tonumber(redis.call('get', key .. ':last') or now)
local delta = math.max(now - last, 0) * rate
tokens = math.min(tokens + delta, burst)
if tokens >= 1 then
    tokens = tokens - 1
    redis.call('set', key, tokens)
    redis.call('set', key .. ':last', now)
    return true
else
    return false
end
`

func (limit *Limit) Allow() bool {

	r, err := redis.NewScript(luaScript).Run(RedisClient, []string{limit.key}, limit.rate, limit.burst, time.Now().Unix()).Bool()
	//r, err := RedisClient.Eval(luaScript, []string{limit.key}, limit.rate, limit.burst, time.Now().Unix()).Bool()
	if err != nil && err != redis.Nil {
		log.Fatalln(err)
	}
	return r
}
