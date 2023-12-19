package requestcontext

import (
	"fmt"
	"log"

	"github.com/dovydasdo/psec/config"
)

type BDProxyAgent struct {
	Config       *config.ConfBDProxy
	CurrentProxy Proxy
	SessionID    int
	Auth         *ProxyAuth
}

func NewBDProxyAgent() *BDProxyAgent {
	c := config.NewBDProxyConf()

	log.Printf("server: %v, un: %v, pwd: %v", c.AuthHost, c.AuthName, c.AuthPass)

	return &BDProxyAgent{
		Config: c,
		Auth: &ProxyAuth{
			Server:   c.AuthHost,
			Username: fmt.Sprintf("%s-session-rand%v", c.AuthName, 0),
			Password: c.AuthPass,
		},
		SessionID: 0,
	}
}

func (p *BDProxyAgent) GetAuth() (*ProxyAuth, error) {
	return p.Auth, nil
}

func (p *BDProxyAgent) SetProxy() error {
	p.SessionID++
	// p.Auth.Username = fmt.Sprintf("%s-session-rand%v", p.Config.AuthName, p.SessionID)
	p.Auth.Username = p.Config.AuthName
	log.Println(p.Auth.Username)
	return nil
}

func (p *BDProxyAgent) LoadProxies() error {
	return nil
}
