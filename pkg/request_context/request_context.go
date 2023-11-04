package requestcontext

import (
	"context"
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
	GetState() *CollectionState
	SetBlockFilter(filter *regexp.Regexp)
	// PerformSimpleRequest(req *http.Request) (*http.Response, error)
}

type DefaultRequestContext struct {
	Page        *rod.Page
	Browser     *rod.Browser
	ProxyAgent  ProxyGetter
	BinPath     string
	HttpClient  *http.Client
	ReqClient   *req.Client
	State       *CollectionState
	StateCheck  interface{}
	CancelF     context.CancelFunc
	Filter      *regexp.Regexp
	BlockFilter *regexp.Regexp
}

func New() *DefaultRequestContext {
	c := &DefaultRequestContext{}
	c.Filter = regexp.MustCompile(".*")
	return c
}

func (r *DefaultRequestContext) GetState() *CollectionState {
	return r.State
}

func (r *DefaultRequestContext) SetBinPath(path string) {
	r.BinPath = path
}

func (r *DefaultRequestContext) Initialize() {
	l := launcher.New()
	if r.BinPath != "" {
		l.Bin(r.BinPath)
	}

	_, err := r.ProxyAgent.GetAuth()
	if err == nil {
		l.Proxy("127.0.0.1:24000")
		log.Println("proxy is set")
	}

	u := l.Leakless(true).Headless(false).MustLaunch()
	r.Browser = rod.New().ControlURL(u).MustConnect()
	r.Page = r.Browser.MustPage("").MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		// UserAgentMetadata: &proto.EmulationUserAgentMetadata{},
	})
	r.ReqClient = req.C().ImpersonateChrome()
	r.State = &CollectionState{}
	r.State.RequestsMade = make([]*proto.NetworkResponseReceived, 0)
	r.SetProxyRouter()
}

func (r *DefaultRequestContext) InitSession() {
	r.Page = r.Browser.MustPage("").MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	})
	r.ReqClient = req.C().ImpersonateChrome()
	r.State = &CollectionState{}
	r.State.RequestsMade = make([]*proto.NetworkResponseReceived, 0)
	r.SetProxyRouter() //test
}

func (r *DefaultRequestContext) PerformRequestInstruction(ins RequestInstruction, ctx *context.Context) (string, error) {
	// for i := 0; i < 1; i++ {
	// 	cctx, cancel := context.WithDeadline(*ctx, time.Now().Add(30*time.Second))
	// 	r.CancelF = cancel

	// 	if ctxVal, ok := rctx.(ReqCtxVal); ok {
	// 		if ctxVal.DoneF == nil {
	// 			//todo add some default behaviour... maybe
	// 			return "", fmt.Errorf("no done function provided, use http client if no complex loading is required")
	// 		}

	// 		r.StateCheck = ctxVal.DoneF
	// 	} else {
	// 		log.Println("failed to get ctx val")
	// 	}

	// 	r.Filter = &ins.Filter

	// 	r.State.LoadState = PROCESSING
	// 	go r.Page.Navigate(ins.URL)
	// 	<-cctx.Done()
	// 	if r.State.LoadState == DONE {
	// 		break
	// 	}

	// 	if r.State.LoadState == PROCESSING {
	// 		log.Println("timedout")
	// 		break
	// 	}

	// 	if r.State.LoadState == FAILED {
	// 		log.Println("blocked")
	// 	}

	// 	r.ProxyAgent.SetProxy()
	// 	r.Page.Close()
	// 	r.HttpClient = nil
	// 	r.InitSession()
	// }

	//Cleanup
	r.Filter = &ins.Filter

	r.Page.Navigate(ins.URL)
	r.Page.MustWaitDOMStable()
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
	// r.SetProxyRouter()
}

func (r *DefaultRequestContext) SetProxyRouter() {
	//todo: pass context to close the router
	router := r.Page.HijackRequests()
	//	defer router.Stop()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		// if r.State.LoadState != PROCESSING {
		// 	ctx.Response.Fail(proto.NetworkErrorReasonAborted)
		// 	return
		// }
		if r.BlockFilter.MatchString(ctx.Request.URL().String()) {
			ctx.Response.Fail(proto.NetworkErrorReasonAborted)
			return
		}

		if r.Filter.MatchString(ctx.Request.URL().String()) {
			log.Println("performing : ", ctx.Request.URL().String())
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
			return
		}

		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})

	go router.Run()
	go r.Page.EachEvent(func(e *proto.NetworkResponseReceived) {

		r.State.RequestsMade = append(r.State.RequestsMade, e)

		// log.Println("req made: ", len(r.State.RequestsMade))
		log.Println("url: ", e.Response.URL)
		if doneF, ok := r.StateCheck.(DoneFunc); ok {
			state := doneF(e.Response, r.State)
			// log.Println("state: ", state)
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
	// go r.Page.EachEvent(func(e *proto.NetworkRequestWillBeSentExtraInfo) {
	// 	log.Println("-------------------------------------------------------------------------")
	// 	for i, h := range e.Headers {
	// 		log.Printf("header key: %v, val: %v \n", i, h)
	// 	}
	// 	log.Println("-------------------------------------------------------------------------")
	// })()
}

func (r *DefaultRequestContext) ChangeProxy() error {
	r.HttpClient = nil
	return r.ProxyAgent.SetProxy()
}

// func (r *DefaultRequestContext) PerformSimpleRequest(req *http.Request) (*http.Response, error) {
// 	if r.HttpClient == nil {
// 		if r.ProxyAgent != nil {
// 			if p, ok := r.ProxyAgent.(*PSECProxyAgent); ok {
// 				r.ReqClient.SetProxyURL(fmt.Sprintf("http://%v", p.CurrentProxy.Ip))
// 			}
// 			if _, ok := r.ProxyAgent.(*BDProxyAgent); ok {
// 				r.ReqClient.SetProxyURL("http://127.0.0.1:24000")
// 			}
// 		}
// 		r.HttpClient = &http.Client{
// 			Transport: r.ReqClient.Transport,
// 		}
// 	}

// 	resp, err := r.ReqClient.R().Get(req.URL.String())

//		return resp.Response, err
//	}
func (r *DefaultRequestContext) SetBlockFilter(filter *regexp.Regexp) {
	r.BlockFilter = filter
}
