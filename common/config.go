package common

import (
	"encoding/json"
	"log"
	"os"
)

const Inventory = 100

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

type Config struct {
	Inventory int64  `json:"inventory"`
	Key       string `json:"key"`
	Rate      int    `json:"rate"`
	Burst     int    `json:"burst"`

	RedCfg  *RedisConfig `json:"redis"`
	MqCfg   *AMQPConfig  `json:"rabbitmq"`
	HttpCfg *HttpConfig  `json:"http"`
}

func InitMqConfig(configFile string) *AMQPConfig {
	mqCfg := &AMQPConfig{
		QueName: "peadx",
		MqUrl:   "amqp://guest:guest@miaosha.peadx.live:5672/miaosha",
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
		Inventory: Inventory,
		Key:       "miaosha_limit",
		Rate:      1,
		Burst:     100,
		RedCfg: &RedisConfig{
			Address:  "miaosha.peadx.live",
			Port:     "6379",
			Password: "123456",
		},
		MqCfg:   &AMQPConfig{QueName: "peadx", MqUrl: "amqp://guest:guest@miaosha.peadx.live:5672/miaosha"},
		HttpCfg: &HttpConfig{Port: "8080"},
	}
	if configFile == "" {
		return config
	}
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return config
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		log.Fatalln(err)
	}
	return config
}
