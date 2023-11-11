package requestcontext

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	util "github.com/dovydasdo/psec/util/injections"
)

type CDPContext struct {
	ctx    context.Context
	cancel context.CancelFunc

	allocator       context.Context
	allocatorCancel context.CancelFunc

	binPath string

	State      *State
	ProxyAgent ProxyGetter
}

func GetCDPContext() *CDPContext {
	// conf := config.NewBrowserlessConf()

	opts := append(chromedp.DefaultExecAllocatorOptions[:]) // chromedp.ProxyServer(fmt.Sprintf("%s:%v", conf.Proxy.Address, conf.Proxy.Port)),

	allocatorContext, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	cdpCtx, cf := chromedp.NewContext(allocatorContext)
	return &CDPContext{
		ctx:             cdpCtx,
		cancel:          cf,
		allocator:       allocatorContext,
		allocatorCancel: cancel,
		State: &State{
			NetworkEvents: make(map[string]*NetworkEvent, 0),
		},
	}
}

func (c *CDPContext) Initialize() {
	// Capture network traffic and save to internal state
	chromedp.ListenTarget(c.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:

			// If there is a response code and status then its a response, let redirects through
			if ev.ResponseStatusCode != 0 && ev.ResponseStatusText != "" {
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
				if event, ok := c.State.NetworkEvents[string(ev.RequestID)]; ok && ev.ResponseStatusCode != 301 && ev.ResponseStatusCode != 302 {
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
								log.Printf("failed to get body for request: %v, err= %v", ev.Request.URL, err.Error())
								time.Sleep(time.Second)
								continue
							}
							event.Response = resp

							return
						}

						log.Printf("after three tries, failed to get body for request: %v", ev.Request.URL)
					}()

				} else {
					log.Printf("got a response when there was no request in the state: %v", ev.Request.URL)
				}

				return
			}
			// Continue with request porcessing
			log.Printf("event request paused: url=%v", ev.Request.URL)
			req := NetworkRequest{}
			req.Body = ev.Request.PostData
			req.URL = ev.Request.URL
			req.Headers = GetHeaders(ev.Request.Headers)

			c.State.NetworkEvents[string(ev.RequestID)] = &NetworkEvent{
				Request: req,
			}

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
	injection, err := GetInjection("/home/ddom/coding/psec/util/injections/stealth.min.js")
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
		fetch.Enable(),
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

	c.State = &State{
		NetworkEvents: make(map[string]*NetworkEvent, 0),
	}
	c.ctx = cdpCtx
	c.cancel = cf
	c.Initialize()
}

func (c *CDPContext) Do(ins ...interface{}) (string, error) {
	for _, instruction := range ins {
		switch v := instruction.(type) {
		case NavigateInstruction:
			url := v.URL
			done, err := GetDoneAction(v.DoneCondition)
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
	return c.State
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

func GetDoneAction(condition interface{}) (chromedp.Action, error) {
	switch c := condition.(type) {
	case DoneElVisible:
		return chromedp.WaitVisible(c), nil
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
