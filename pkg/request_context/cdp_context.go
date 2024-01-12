package requestcontext

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
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

		case *fetch.EventRequestPaused:

			// If there is a response code and status then its a response, let redirects through
			// TODO: look for a better way to distinguish between requests and responses
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
				// Process Response
				if event, ok := c.State.NetworkEvents.Load(ev.RequestID); ok && ev.ResponseStatusCode != 301 && ev.ResponseStatusCode != 302 {
					resp := NetworkResponse{}

					go func() {

						for i := 0; i < 3; i++ {
							err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
								body, err := fetch.GetResponseBody(ev.RequestID).Do(ctx)
								resp.Body = string(body)
								return err
							}))

							resp.URL = ev.Request.URL
							resp.Headers = GetHeadersResp(ev.ResponseHeaders)

							if err != nil {
								c.logger.Error("cdp", "message", "failed to get body for request", "error", err.Error())
								time.Sleep(time.Second)
								continue
							}

							if e, ok := event.(*NetworkEvent); ok {
								e.Response = resp
							}

							c.logger.Log(context.Background(), -1, "RESPONSE", "url", ev.Request.URL, "status", ev.ResponseStatusCode, "id", ev.RequestID)

							return
						}

						c.logger.Warn("cdp", "message", "after three tries, failed to get body for request", "url", ev.Request.URL)
					}()

				} else {
					c.logger.Warn("cdp", "message", "got a response when there was no request in the statev", "url", ev.Request.URL)
				}

				return
			}
			c.logger.Log(context.Background(), -1, "REQUEST", "url", ev.Request.URL, "id", ev.RequestID)
			// Continue with request porcessing
			req := NetworkRequest{}
			req.Body = ev.Request.PostData
			req.URL = ev.Request.URL
			req.Headers = GetHeaders(ev.Request.Headers)

			c.State.NetworkEvents.Store(ev.RequestID, &NetworkEvent{Request: req})

			// Let request pass
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
		}
	})

	// Todo: make configurable
	injection, err := GetInjection(c.Config.InjectionPath)
	if err != nil {
		log.Fatalf("failed to read injection, aborting. Err: %v", err.Error())
	}

	addInjection := page.AddScriptToEvaluateOnNewDocument(injection)
	addInjection.RunImmediately = true

	uaInfo := util.GetStaticUAInfo()
	overrideUA := emulation.SetUserAgentOverride(uaInfo.UserAgent)
	overrideUA.AcceptLanguage = uaInfo.AcceptLanguage
	overrideUA.Platform = uaInfo.Platform
	overrideUA.UserAgentMetadata = &emulation.UserAgentMetadata{
		Architecture:    uaInfo.Metadata.Architecture,
		Bitness:         uaInfo.Metadata.Bitness,
		Mobile:          uaInfo.Metadata.Mobile,
		Model:           uaInfo.Metadata.Model,
		Platform:        uaInfo.Metadata.Platform,
		PlatformVersion: uaInfo.Metadata.PlatformVersion,
		Wow64:           uaInfo.Metadata.WOW64,
		Brands: []*emulation.UserAgentBrandVersion{
			{
				Brand:   uaInfo.Metadata.Brands[0].Brand,
				Version: uaInfo.Metadata.Brands[0].Version,
			},
			{
				Brand:   uaInfo.Metadata.Brands[1].Brand,
				Version: uaInfo.Metadata.Brands[1].Version,
			},
			{
				Brand:   uaInfo.Metadata.Brands[2].Brand,
				Version: uaInfo.Metadata.Brands[2].Version,
			},
		},
		FullVersionList: []*emulation.UserAgentBrandVersion{
			{
				Brand:   uaInfo.Metadata.FullVersionList[0].Brand,
				Version: uaInfo.Metadata.FullVersionList[0].Version,
			},
			{
				Brand:   uaInfo.Metadata.FullVersionList[1].Brand,
				Version: uaInfo.Metadata.FullVersionList[1].Version,
			},
			{
				Brand:   uaInfo.Metadata.FullVersionList[2].Brand,
				Version: uaInfo.Metadata.FullVersionList[2].Version,
			},
		},
	}

	overrideAutomation := emulation.SetAutomationOverride(false)

	chromedp.Run(c.ctx,
		fetch.Enable().WithHandleAuthRequests(true),
		network.Enable(),
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
		chromedp.Navigate(""))
}

func (c *CDPContext) Reset() {
	c.cancel()

	cdpCtx, cf := chromedp.NewContext(c.allocator)

	c.State = &State{}
	c.ctx = cdpCtx
	c.cancel = cf
	c.Initialize()
}

func (c *CDPContext) Do(ins ...interface{}) (string, error) {
	for _, instruction := range ins {
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

			if err != nil {
				return "", err
			}

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

	return html, err
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
			for i := 0; i < 20; i++ {
				c.State.NetworkEvents.Range(func(key, value any) bool {
					if v, ok := value.(*NetworkEvent); ok {
						if v.Response.URL == string(cond) {
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
		return nil, errors.New("the provided contition was not recognised")
	}
}

func GetHeaders(protoHeaders network.Headers) map[string]string {
	hto := make(map[string]string, 0)
	for hName, hVal := range protoHeaders {
		if val, ok := hVal.(string); ok {
			hto[hName] = val
			continue
		}

		log.Printf("some funky header: %v", hName)
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
