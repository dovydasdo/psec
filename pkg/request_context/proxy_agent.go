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
	//TODO: implement some profiles per site for more reasonable proxy handling
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
	log.Println(responseBody)
	return nil
}

func (a *PSECProxyAgent) SetProxy() error {
	for key, value := range a.ActiveProxies {
		a.CurrentProxy = value
		delete(a.ActiveProxies, key)
		return nil
	}
	return errors.New("no proxies")
}
