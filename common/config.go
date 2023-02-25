package common

import (
	"encoding/json"
	"log"
	"os"
)

type RedisConfig struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

type AMQPConfig struct {
	MqUrl string `json:"mqUrl"`
}

type GinConfig struct {
	Port string `json:"port"`
}

type Config struct {
	RedCfg RedisConfig `json:"redis"`
	MqCfg  AMQPConfig  `json:"rabbitmq"`
	GinCfg GinConfig   `json:"gin"`
}

func InitConfig(configFile string) *Config {
	config := &Config{
		RedCfg: RedisConfig{
			Address:  "miaosha.peadx.live",
			Port:     "6379",
			Password: "123456",
		},
		MqCfg:  AMQPConfig{MqUrl: "amqp://guest:guest@miaosha.peadx.live:5672/miaosha"},
		GinCfg: GinConfig{Port: "8080"},
	}
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Println(err)
		return config
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		log.Fatalln(err)
	}
	return config
}
