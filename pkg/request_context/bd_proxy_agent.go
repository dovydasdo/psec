package requestcontext

import (
	"log"
)

type BDProxyAgent struct {
	CurrentProxy Proxy
	SessionID    int
	Auth         *ProxyAuth
}

func NewBDProxyAgent(opts *BDProxyOptions) *BDProxyAgent {
	return &BDProxyAgent{
		Auth: &ProxyAuth{
			Server:   opts.Server,
			Username: opts.Username,
			Password: opts.Password,
		},
		SessionID: 0,
	}
}

func (p *BDProxyAgent) GetAuth() (*ProxyAuth, error) {
	return p.Auth, nil
}

func (p *BDProxyAgent) SetProxy() error {
	p.SessionID++
	// TODO: make username id increment on new proxy request
	log.Println(p.Auth.Username)
	return nil
}

func (p *BDProxyAgent) LoadProxies() error {
	return nil
}
