package requestcontext

type ProxyGetter interface {
	GetProxies([]Proxy, error)
	//TODO: implement some profiles per site for more reasonable proxy handling
}

type PSECProxyAgent struct {
}

func NewPSECProxyAgent() *PSECProxyAgent {
	return &PSECProxyAgent{}
}

func (a *PSECProxyAgent) GetProxies() ([]Proxy, error) {

	return nil, nil
}

func (a *PSECProxyAgent) getProxiesFromAPI() {
	//TODO
}
