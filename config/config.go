// Package config 总初始化文件
package config

import (
	_ "embed"

	"go.yaml.in/yaml/v2"
)

// 总配文件
type config struct {
	Server   Server   `yaml:"server"`
	Mysql    Mysql    `yaml:"mysql"`
	Redis    Redis    `yaml:"redis"`
	Rocket   Rocket   `yaml:"rocket"`
	ES       ES       `yaml:"es"`
	Template Template `yaml:"template"`
}

// key-pair
type keyPair struct {
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
	HMACKey    string `yaml:"HMAC_key"`
}

// 项目端口配置
type Server struct {
	Port       string `yaml:"port"`
	Address    string `yaml:"address"`
	Model      string `yaml:"model"`
	ServerName string `yaml:"serverName"`
}

// 数据库配置
type Mysql struct {
	Dialects string `yaml:"dialects"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Db       string `yaml:"db"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"`
	MaxOpen  int    `yaml:"maxOpen"`
}

// Redis配置
type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

type ES struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

// Log日志
type Log struct {
	Path  string `yaml:"path"`
	Name  string `yaml:"name"`
	Model string `yaml:"model"`
}

type redisCluster struct {
	Addresses []string `yaml:"addresses"`
	Password  string   `yaml:"password"`
	Db        int      `yaml:"db"`
	PoolSize  int      `yaml:"poolSize"`
}

type Rocket struct {
	NameServer       string `yaml:"nameServer"`
	ProducerGroup    string `yaml:"producer_group"`
	ConsumerGroup    string `yaml:"consumer_group"`
	TransactionGroup string `yaml:"pransaction_group"`
}

type Template struct {
	MinCount  int    `yaml:"minCount"`
	MaxCount  int    `yaml:"maxCount"`
	VideoRoot string `yaml:"videoRoot"`
}

//go:embed config.yaml
var configData []byte

var Config = &config{}

// 配置初始化 改成手动调用
func init() {
	// 绑定值
	err := yaml.Unmarshal(configData, Config)
	if err != nil {
		panic(err)
	}
}
