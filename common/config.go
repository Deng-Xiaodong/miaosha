package common

import (
	"encoding/json"
	"log"
	"os"
)

const Inventory = 10 << 10

type RedisConfig struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

type AMQPConfig struct {
	QueName string `json:"queName"`
	MqUrl   string `json:"mqUrl"`
}

type HttpConfig struct {
	Port string `json:"port"`
}

type LimitConfig struct {
	Inventory int64  `json:"inventory"`
	LimitKey  string `json:"limitKey"`
	Burst     int64  `json:"burst"` //当前令牌数量，初始化为最大值burst
	Rate      int64  `json:"rate"`
}
type Config struct {
	LimitCfg *LimitConfig `json:"limit"`
	RedCfg   *RedisConfig `json:"redis"`
	MqCfg    *AMQPConfig  `json:"rabbitmq"`
	HttpCfg  *HttpConfig  `json:"http"`
}

func InitMqConfig(configFile string) *AMQPConfig {
	mqCfg := &AMQPConfig{
		QueName: "peadx",
		MqUrl:   "amqp://root:root@rabbitmq1:5672/miaosha",
	}
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return mqCfg
	}
	if err := json.Unmarshal(bytes, mqCfg); err != nil {
		log.Fatalln(err)
	}
	return mqCfg
}

func InitConfig(configFile string) *Config {
	//默认配置为本地测试环境
	config := &Config{
		LimitCfg: &LimitConfig{
			Inventory: Inventory,
			LimitKey:  "miaosha_limit",
			Rate:      1,
			Burst:     100,
		},
		RedCfg: &RedisConfig{
			Address:  "redis",
			Port:     "6379",
			Password: "123456",
		},
		MqCfg:   &AMQPConfig{QueName: "peadx", MqUrl: "amqp://root:root@rabbitmq1:5672/miaosha"},
		HttpCfg: &HttpConfig{Port: "8080"},
	}
	if configFile == "" {
		return config
	}
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("文件读取错误\n")
		return config
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		log.Fatalln(err)
	}
	return config
}
