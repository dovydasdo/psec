package config

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	DB ConfDB
}

type ConfDB struct {
	Host     string `env:"DB_HOST,required"`
	Port     int    `env:"DB_PORT,required"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASS,required"`
	DBName   string `env:"DB_NAME,required"`
	Debug    bool   `env:"DB_DEBUG,required"`
}

func New() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

type ProxyConf struct {
	Address string `env:"PROXY_API_ADDRESS, required"`
	Port    int    `env:"PROXY_API_PORT, required"`
}

func NewProxyConfig() *ProxyConf {
	var c ProxyConf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
