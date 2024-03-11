package config

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	DB         ConfDB
	NATSFile   string `env:"NATS_CREDS"`
	NATSServer string `env:"NATS_SERVER"`
}

type ConfDB struct {
	Host     string `env:"DB_HOST,required"`
	Port     int    `env:"DB_PORT,required"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASS,required"`
	DBName   string `env:"DB_NAME,required"`
	Debug    bool   `env:"DB_DEBUG,required"`
}

type ProxyConf struct {
	Address  string `env:"PROXY_API_ADDRESS, required"`
	Port     int    `env:"PROXY_API_PORT, required"`
	AuthName string `env:"PROXY_AUTH_NAME"`
	AuthHost string `env:"PROXY_AUTH_HOST"`
	AuthPass string `env:"PROXY_AUTH_PASS"`
}

type ConfTurso struct {
	DBName  string `env:"TURSO_DB,required"`
	DBToken string `env:"TURSO_TOKEN,required"`
	Debug   bool   `env:"DB_DEBUG,required"`
}

type ConfBrowserless struct {
	Token string `env:"BROWSERLESS_TOKEN,required"`
	Proxy ProxyConf
}

type ConfCDPLaunch struct {
	Proxy         ProxyConf
	BinPath       string `env:"CDP_BIN_PATH,required"`
	InjectionPath string `env:"INJECTION_PATH,required"`
}

type ConfBDProxy struct {
	AuthName string `env:"PROXY_AUTH_NAME"`
	AuthHost string `env:"PROXY_AUTH_HOST"`
	AuthPass string `env:"PROXY_AUTH_PASS"`
}

func New() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewProxyConfig() *ProxyConf {
	var c ProxyConf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewTursoConf() *ConfTurso {
	var c ConfTurso
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewBrowserlessConf() *ConfBrowserless {
	var c ConfBrowserless
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewCDPLaunchConf() *ConfCDPLaunch {
	var c ConfCDPLaunch
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewBDProxyConf() *ConfBDProxy {
	var c ConfBDProxy
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
