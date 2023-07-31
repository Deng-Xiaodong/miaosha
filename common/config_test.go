package common

import (
	"fmt"
	"testing"
)

func TestInitConfig(t *testing.T) {
	config := InitConfig("config.json")

	fmt.Printf("config:%v", config.LimitCfg)
}
