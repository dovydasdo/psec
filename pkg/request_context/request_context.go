package requestcontext

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type RequestContextInterface interface {
	PerformRequestInstruction(ins RequestInstruction) error
	GetRequestAgent() (interface{}, error)
	RegisterProxyAgent(a ProxyGetter)
	SetBinPath(path string)
	Initialize()
}

type DefaultRequestContext struct {
	Page       *rod.Page
	ProxyAgent ProxyGetter
	BinPath    string
	//TODO: some prox implmentation
}

func New() *DefaultRequestContext {
	c := &DefaultRequestContext{}
	return c
}

func (r *DefaultRequestContext) SetBinPath(path string) {
	r.BinPath = path
}

func (r *DefaultRequestContext) Initialize() {
	l := launcher.New()
	if r.BinPath != "" {
		l.Bin(r.BinPath)
	}

	if r.ProxyAgent != nil {
		l.Proxy(r.ProxyAgent.GetProxy())
	}

	u := l.Leakless(true).Headless(true).MustLaunch()
	r.Page = rod.New().ControlURL(u).MustConnect().MustPage("")
}

func (r *DefaultRequestContext) PerformRequestInstruction(ins RequestInstruction) error {
	r.Page.Navigate(ins.URL)
	r.Page.MustWaitDOMStable()
	return nil
}

func (r *DefaultRequestContext) GetRequestAgent() (interface{}, error) {
	return r.Page, nil
}

func (r *DefaultRequestContext) RegisterProxyAgent(a ProxyGetter) {
	r.ProxyAgent = a
	a.LoadProxies()
}
