package requestcontext

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/dovydasdo/psec/config"
)

type ProxyGetter interface {
	LoadProxies() error
	GetProxy() string
	//TODO: implement some profiles per site for more reasonable proxy handling
}

type PSECProxyAgent struct {
	APIAddress    string
	APIPort       int
	ActiveProxies Proxies
	Config        *config.ProxyConf
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

	//todo: make a call to the api and demarshall the proxies

	a.ActiveProxies = proxies
	log.Println(responseBody)
	return nil
}

func (a *PSECProxyAgent) GetProxy() string {
	for _, value := range a.ActiveProxies {
		return value.Ip
	}
	return ""
}
