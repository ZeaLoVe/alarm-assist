package g

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type QueueConfig struct {
	Sms   string `json:"sms"`
	Mail  string `json:"mail"`
	IMSms string `json:"im"`
	Phone string `json:"phone"`
}

type RedisConfig struct {
	Addr        string   `json:"addr"`
	MaxIdle     int      `json:"maxIdle"`
	MaxConsumer int      `json:"maxConsumer"`
	QueryQueues []string `json:"queryQueues"`
}

//group用来记录发送历史，为99U群号，所有发送历史都会汇总发到该群组（通过逗号分割）
type IMConfig struct {
	Uri        string `json:"uri"`
	Password   string `json:"password"`
	Group      string `json:"group,omitempty"`
	PhoneGroup string `json:"phonegroup,omitempty"`
}

type GlobalConfig struct {
	Debug  bool         `json:"debug"`
	Portal string       `json:"portal"`
	Uic    string       `json:"uic"`
	Http   *HttpConfig  `json:"http"`
	Queue  *QueueConfig `json:"queue"`
	Redis  *RedisConfig `json:"redis"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	configLock.Lock()
	defer configLock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
}
