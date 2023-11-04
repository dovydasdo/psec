package requestcontext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/dovydasdo/psec/config"
)

type ProxyGetter interface {
	LoadProxies() error
	SetProxy() error
	GetAuth() (*ProxyAuth, error)
}

type ProxyAuth struct {
	Server   string
	Username string
	Password string
}

type PSECProxyAgent struct {
	APIAddress    string
	APIPort       int
	ActiveProxies Proxies
	Config        *config.ProxyConf
	CurrentProxy  Proxy
}

func NewPSECProxyAgent() *PSECProxyAgent {
	c := config.NewProxyConfig()
	return &PSECProxyAgent{
		Config: c,
	}
}

func (a *PSECProxyAgent) GetAuth() (*ProxyAuth, error) {
	return nil, fmt.Errorf("defautl agent has no auth")
}

func (a *PSECProxyAgent) LoadProxies() error {
	endpoint := fmt.Sprintf("http://%v:%v/proxies", a.Config.Address, a.Config.Port)

	proxies := make(Proxies)

	response, err := http.Get(endpoint)
	if err != nil {
		return err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, &proxies); err != nil {
		return err
	}

	a.ActiveProxies = proxies
	log.Println(string(responseBody[:]))
	return nil
}

func (a *PSECProxyAgent) SetProxy() error {
	for key, value := range a.ActiveProxies {
		a.CurrentProxy = value
		log.Println("current pxy: ", a.CurrentProxy.Ip)
		delete(a.ActiveProxies, key)
		return nil
	}
	return errors.New("no proxies")
}
