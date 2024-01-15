package requestcontext

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/dovydasdo/psec/config"
	util "github.com/dovydasdo/psec/util/injections"
)

type CDPContext struct {
	ctx    context.Context
	cancel context.CancelFunc

	allocator       context.Context
	allocatorCancel context.CancelFunc

	binPath string
	logger  *slog.Logger

	State      *State
	ProxyAgent ProxyGetter
	Config     config.ConfCDPLaunch
}

type Result struct {
	Name     string
	Duration time.Duration
	Value    interface{}
	Type     string
	Error    error
}

func GetCDPContext(conf config.ConfCDPLaunch, l *slog.Logger) *CDPContext {
	return &CDPContext{
		State:  &State{},
		logger: l,
		Config: conf,
	}
}

func (c *CDPContext) Initialize() {
	// if proxy agent has been registered set the proxy
	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.Flag("ignore-certificate-errors", true))

	if bdAgent, ok := c.ProxyAgent.(*BDProxyAgent); ok {
		proxyConf := bdAgent.Config
		opts = append(opts, chromedp.ProxyServer(fmt.Sprintf("http://%v", proxyConf.AuthHost)))
	}

	opts = append(opts, chromedp.ExecPath(c.Config.BinPath))

	allocatorContext, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	cdpCtx, cf := chromedp.NewContext(allocatorContext)

	c.binPath = c.Config.BinPath
	c.cancel = cf
	c.ctx = cdpCtx
	c.allocator = allocatorContext
	c.allocatorCancel = cancel

	// Capture network traffic and save to internal state
	chromedp.ListenTarget(c.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventAuthRequired:
			if ev.AuthChallenge.Source == fetch.AuthChallengeSourceProxy {
				go func() {
					auth, err := c.ProxyAgent.GetAuth()
					if err != nil {
						log.Fatal(err)
					}

					_ = chromedp.Run(c.ctx,
						fetch.ContinueWithAuth(ev.RequestID, &fetch.AuthChallengeResponse{
							Response: fetch.AuthChallengeResponseResponseProvideCredentials,
							Username: auth.Username,
							Password: auth.Password,
						}),
					)
				}()
			}
		case *network.EventResponseReceived:
			if event, ok := c.State.NetworkEvents.Load(ev.RequestID); ok {
				c.logger.Debug("cdp", "received", ev.Response.URL)

				resp := NetworkResponse{}
				go func() {
					c.logger.Debug("cdp", "getting body for url", ev.Response.URL)

					err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
						body, err := network.GetResponseBody(ev.RequestID).Do(ctx)
						resp.Body = string(body)
						return err
					}))

					resp.URL = ev.Response.URL

					if err != nil {
						time.Sleep(time.Second)
					}

					if e, ok := event.(*NetworkEvent); ok {
						e.Response = resp
					}
				}()
			}

		case *network.EventRequestWillBeSent:
			c.logger.Debug("cdp", "request will be sent", ev.Request.URL)

			c.logger.Log(context.Background(), -1, "REQUEST", "url", ev.Request.URL, "id", ev.RequestID)
			// Continue with request porcessing
			req := NetworkRequest{}
			req.Body = ev.Request.PostData
			req.URL = ev.Request.URL
			req.Headers = GetHeaders(ev.Request.Headers)

			c.State.NetworkEvents.Store(ev.RequestID, &NetworkEvent{Request: req})

		case *fetch.EventRequestPaused:
			// If there is a response code and status then its a response, let redirects through
			// TODO: look for a better way to distinguish between requests and responses
			if ev.ResponseStatusCode == 301 || ev.ResponseStatusCode == 302 {
				go func() {
					err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
						params := fetch.ContinueRequest(ev.RequestID)
						params.InterceptResponse = true
						return params.Do(ctx)
					}))

					if err != nil {
						log.Println(err.Error())
					}
				}()
				return
			}
			if ev.ResponseStatusCode != 0 {
				// Continue the response
				go func() {
					err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
						params := fetch.ContinueResponse(ev.RequestID)
						return params.Do(ctx)
					}))

					if err != nil {
						log.Println(err.Error())
					}
				}()

				return
			}

			// Let request pass
			go func() {
				err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					params := fetch.ContinueRequest(ev.RequestID)
					params.InterceptResponse = true
					return params.Do(ctx)
				}))

				c.logger.Debug("cdp", "continue done url", ev.Request.URL)

				if err != nil {
					log.Println(err.Error())
				}
			}()
		}
	})

	injection, err := GetInjection(c.Config.InjectionPath)
	if err != nil {
		log.Fatalf("failed to read injection, aborting. Err: %v", err.Error())
	}

	addInjection := page.AddScriptToEvaluateOnNewDocument(injection)
	addInjection.RunImmediately = true

	uaInfo, err := util.GetLatestUAInfo()
	if err != nil {
		c.logger.Warn("cdp", "message", "failed to get latest user agent data, static data will be used")
	}

	overrideUA := emulation.SetUserAgentOverride(uaInfo.UserAgent)
	overrideUA.AcceptLanguage = uaInfo.AcceptLanguage
	overrideUA.Platform = uaInfo.Platform
	overrideUA.UserAgentMetadata = &emulation.UserAgentMetadata{
		Architecture:    uaInfo.Metadata.JsHighEntropyHints.Architecture,
		Bitness:         uaInfo.Metadata.JsHighEntropyHints.Bitness,
		Mobile:          uaInfo.Metadata.JsHighEntropyHints.Mobile,
		Model:           uaInfo.Metadata.JsHighEntropyHints.Model,
		Platform:        uaInfo.Metadata.JsHighEntropyHints.Platform,
		PlatformVersion: uaInfo.Metadata.JsHighEntropyHints.PlatformVersion,
		Wow64:           uaInfo.Metadata.JsHighEntropyHints.Wow64,
		Brands: []*emulation.UserAgentBrandVersion{
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.Brands[0].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.Brands[0].Version,
			},
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.Brands[1].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.Brands[1].Version,
			},
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.Brands[2].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.Brands[2].Version,
			},
		},
		FullVersionList: []*emulation.UserAgentBrandVersion{
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.FullVersionList[0].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.FullVersionList[0].Version,
			},
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.FullVersionList[1].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.FullVersionList[1].Version,
			},
			{
				Brand:   uaInfo.Metadata.JsHighEntropyHints.FullVersionList[2].Brand,
				Version: uaInfo.Metadata.JsHighEntropyHints.FullVersionList[2].Version,
			},
		},
	}

	overrideAutomation := emulation.SetAutomationOverride(false)

	err = chromedp.Run(c.ctx,
		network.Enable(),
		fetch.Enable().WithHandleAuthRequests(true),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Fingerprint stuff
			_, err := addInjection.Do(ctx)
			if err != nil {
				return err
			}

			err = overrideUA.Do(ctx)
			if err != nil {
				return err
			}

			err = overrideAutomation.Do(ctx)
			if err != nil {
				return err
			}

			return err
		}),
		chromedp.Navigate("about:blank"))

	if err != nil {
		log.Fatalf("failed to start chromedp %v", err)
	}
}

func (c *CDPContext) Reset() {
	c.cancel()

	cdpCtx, cf := chromedp.NewContext(c.allocator)

	c.State = &State{}
	c.ctx = cdpCtx
	c.cancel = cf
	c.Initialize()
}

func (c *CDPContext) Close() {
	c.cancel()
}

func (c *CDPContext) Do(ins ...interface{}) ([]Result, error) {
	result := make([]Result, 0)

	doStart := time.Now()

	for _, instruction := range ins {
		insStart := time.Now()
		switch v := instruction.(type) {
		case NavigateInstruction:
			url := v.URL
			done, err := c.GetDoneAction(v.DoneCondition)
			if err != nil {
				continue
			}

			err = chromedp.Run(c.ctx,
				network.SetBlockedURLS(v.Filters),
				chromedp.Navigate(url),
				done,
				network.SetBlockedURLS(make([]string, 0)),
			)

			res := Result{
				Type:     "navigate",
				Duration: time.Now().Sub(insStart),
				Error:    err,
			}

			result = append(result, res)

			if err != nil {
				return result, err
			}
		case JSEvalInstruction:
			script := v.Script
			_ = v.Timeout // not used for now
			res := Result{
				Type:     "js_eval",
				Duration: time.Now().Sub(insStart),
			}

			err := chromedp.Run(c.ctx,
				runtime.Enable(),
				chromedp.Evaluate(script, v.Result),
			)

			// this is stupid
			res.Value = v.Result

			res.Error = err
			result = append(result, res)
		case string:
			log.Println(v)
			continue
		default:
			log.Println("Instruction type not recognised, skipping")
			continue
		}
	}

	// Todo: consider of some more fancy return types are neede
	var html string
	err := chromedp.Run(c.ctx,
		chromedp.Evaluate(`document.documentElement.outerHTML`, &html),
	)

	result = append(result, Result{Type: "html", Value: html, Duration: time.Now().Sub(doStart)})
	return result, err
}

func (c *CDPContext) Cancel() {
	c.cancel()
}

func (c *CDPContext) GetState() *State {
	html := ""
	chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	}))
	c.State.Source = html
	return c.State
}

func (c *CDPContext) ClearState() {
	c.State = &State{}
}

func (c *CDPContext) SetBinPath(path string) {
	c.binPath = path
}

func (c *CDPContext) RegisterProxyAgent(a ProxyGetter) {
	c.ProxyAgent = a
}

func (c *CDPContext) ChangeProxy() error {
	return c.ProxyAgent.SetProxy()
}

func (c CDPContext) GetDoneAction(condition interface{}) (chromedp.Action, error) {
	switch cond := condition.(type) {
	case DoneElVisible:
		return chromedp.WaitVisible(cond), nil
	case DoneResponseReceived:
		return chromedp.ActionFunc(func(ctx context.Context) error {
			//TODO: implement without polling
			toBreak := false
			cURL, err := url.Parse(string(cond))
			if err != nil {
				return nil
			}
			for i := 0; i < 20; i++ {
				c.State.NetworkEvents.Range(func(key, value any) bool {
					if v, ok := value.(*NetworkEvent); ok {
						vURL, err := url.Parse(v.Response.URL)
						if err != nil {
							return true
						}

						if normalizeURL(vURL) == normalizeURL(cURL) {
							toBreak = true
							return true
						}
					}

					return true
				})

				if toBreak {
					break
				}

				time.Sleep(time.Millisecond * 500)
			}

			return nil
		}), nil
	default:
		return nil, errors.New("the provided condition was not recognised")
	}
}

func GetHeaders(protoHeaders network.Headers) map[string]string {
	hto := make(map[string]string, 0)
	for hName, hVal := range protoHeaders {
		if val, ok := hVal.(string); ok {
			hto[hName] = val
			continue
		}
	}

	return hto
}

func GetHeadersResp(protoHeaders []*fetch.HeaderEntry) map[string]string {
	hto := make(map[string]string, 0)
	for _, entry := range protoHeaders {
		hto[entry.Name] = entry.Value
	}

	return hto
}

func GetInjection(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file[:]), nil
}

func normalizeURL(u *url.URL) string {
	u.Host = strings.ToLower(u.Host)
	if len(u.Path) == 0 {
		return u.String()
	}
	if u.Path[len(u.Path)-1:] == "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}
	return u.String()
}
