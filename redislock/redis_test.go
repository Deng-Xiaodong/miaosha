package redislock

import (
	"fmt"
	"miaosha/common"
	"sync"
	"testing"
)

var limit *Limit

func init() {
	config := common.InitConfig("")
	InitRedisClient(config.RedCfg)
	limit = NewLimit(config.Key, config.Rate, config.Burst)
}
func TestGetOne(t *testing.T) {
	var wg sync.WaitGroup
	//wg.Add(10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			fmt.Printf("%d是否抢到锁:%v\n", id, GetOne())
			GetOne()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
func TestAllow(t *testing.T) {

	for i := 0; i < 20; i++ {
		fmt.Printf("%d是否获取到令牌：%v\n", i, limit.Allow())
	}
}
