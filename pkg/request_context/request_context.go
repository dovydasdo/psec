package requestcontext

import (
	"fmt"
	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/imroc/req/v3"
)

type RequestContextInterface interface {
	PerformRequestInstruction(ins RequestInstruction) error
	GetRequestAgent() (interface{}, error)
	RegisterProxyAgent(a ProxyGetter)
	SetBinPath(path string)
	Initialize()
	ChangeProxy() error
}

type DefaultRequestContext struct {
	Page       *rod.Page
	ProxyAgent ProxyGetter
	BinPath    string
	HttpClient *http.Client
	ReqClient  *req.Client
	State      *CollectionState
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

	u := l.Leakless(true).Headless(true).MustLaunch()
	r.Page = rod.New().ControlURL(u).MustConnect().MustPage("")
	r.ReqClient = req.C().ImpersonateChrome()
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
	a.SetProxy()
	r.SetProxyRouter()
}

func (r *DefaultRequestContext) SetProxyRouter() {
	//todo: pass context to close the router
	router := r.Page.HijackRequests()
	//	defer router.Stop()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		ctx.Request.Type()
		if r.HttpClient == nil {
			// todo: handle proxy changing for same ctx
			if r.ProxyAgent != nil {
				if p, ok := r.ProxyAgent.(*PSECProxyAgent); ok {
					r.ReqClient.SetProxyURL(fmt.Sprintf("http://%v", p.CurrentProxy.Ip))
				}
			}
			r.HttpClient = &http.Client{
				Transport: r.ReqClient.Transport,
			}
		}

		ctx.LoadResponse(r.HttpClient, true)
	})

	go router.Run()
	go r.Page.EachEvent(func(e *proto.NetworkResponseReceived) {
		r.State.RequestsMade = append(r.State.RequestsMade, e.Response)
	})()
}

func (r *DefaultRequestContext) ChangeProxy() error {
	r.HttpClient = nil
	return r.ProxyAgent.SetProxy()
}

// If this fai1s collection should be terminated
func (r *DefaultRequestContext) MustLoad(url string, retryCount int) *http.Response {
	resp, err := r.ReqClient.NewRequest().Get(url)
	if err != nil {
		//restart of something
		return nil
	}

	return resp.Response
}
