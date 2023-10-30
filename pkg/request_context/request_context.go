package requestcontext

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/imroc/req/v3"
)

type RequestContextInterface interface {
	PerformRequestInstruction(ins RequestInstruction, ctx *context.Context) (string, error)
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
	StateCheck interface{}
	CancelF    context.CancelFunc
	Filter     *regexp.Regexp
}

func New() *DefaultRequestContext {
	c := &DefaultRequestContext{}
	c.Filter = regexp.MustCompile(".*")
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

	u := l.Leakless(true).Headless(false).MustLaunch()
	r.Page = rod.New().ControlURL(u).MustConnect().MustPage("")
	r.ReqClient = req.C().ImpersonateChrome()
	r.State = &CollectionState{}
	r.State.RequestsMade = make([]*proto.NetworkResponse, 0)
	r.SetProxyRouter() //test
}

func (r *DefaultRequestContext) PerformRequestInstruction(ins RequestInstruction, ctx *context.Context) (string, error) {
	key := ReqCtxKey{Id: ins.URL}

	rctx := (*ctx).Value(key)

	cctx, cancel := context.WithCancel(*ctx)
	r.CancelF = cancel

	if ctxVal, ok := rctx.(ReqCtxVal); ok {
		if ctxVal.DoneF == nil {
			//todo add some default behaviour... maybe
			return "", fmt.Errorf("no done function provided, use http client if no complex loading is required")
		}

		r.StateCheck = ctxVal.DoneF
	} else {
		log.Println("failed to get ctx val")
	}

	r.Filter = &ins.Filter

	r.State.LoadState = PROCESSING
	r.Page.Navigate(ins.URL)
	<-cctx.Done()

	//Cleanup
	r.CancelF = nil
	r.StateCheck = nil
	r.Filter = regexp.MustCompile(".*")

	return r.Page.HTML()
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
			if r.ProxyAgent != nil {
				if p, ok := r.ProxyAgent.(*PSECProxyAgent); ok {
					r.ReqClient.SetProxyURL(fmt.Sprintf("http://%v", p.CurrentProxy.Ip))
				}
			}
			r.HttpClient = &http.Client{
				Transport: r.ReqClient.Transport,
			}
		}

		if r.State.LoadState != PROCESSING {
			ctx.Response.Fail(proto.NetworkErrorReasonAborted)
			return
		}

		if r.Filter.MatchString(ctx.Request.URL().String()) {
			log.Println("performing : ", ctx.Request.URL().String())
			ctx.LoadResponse(r.HttpClient, true)
		} else {
			log.Println(r.Filter.String(), " === ", ctx.Request.URL())
			log.Println("failed to match regex")
		}
	})

	go router.Run()
	go r.Page.EachEvent(func(e *proto.NetworkResponseReceived) {
		r.State.RequestsMade = append(r.State.RequestsMade, e.Response)

		// log.Println("req made: ", len(r.State.RequestsMade))
		// log.Println("url: ", e.Response.URL)
		if doneF, ok := r.StateCheck.(DoneFunc); ok {
			state := doneF(e.Response, r.State)
			switch state {
			case SUCCESS:
				r.State.LoadState = DONE
				r.CancelF()
			case TIMEOUT:
				r.State.LoadState = FAILED
				r.CancelF()
			case BLOCKED:
				r.State.LoadState = FAILED
				r.CancelF()
			case CONTINUE:
				r.State.LoadState = PROCESSING
			default:
				log.Println("state not recognised, bad...")
			}

		}
	})()
}

func (r *DefaultRequestContext) ChangeProxy() error {
	r.HttpClient = nil
	return r.ProxyAgent.SetProxy()
}
