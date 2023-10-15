package requestcontext

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type RequestContextInterface interface {
	PerformRequestInstruction(ins RequestInstruction) error
	GetRequestAgent() (interface{}, error)
}

type DefaultRequestContext struct {
	Page *rod.Page
	//TODO: some prox implmentation
}

func New() *DefaultRequestContext {
	c := &DefaultRequestContext{}
	c.Initialize()
	return c
}

func (r *DefaultRequestContext) Initialize() {
	u := launcher.New().Leakless(true).Headless(false).MustLaunch()
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
