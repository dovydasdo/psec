package requestcontext

import "github.com/go-rod/rod"

type RequestContextInterface interface {
	PerformRequestInstruction(ins RequestInstruction) error
	GetRequestAgent() (interface{}, error)
}

type DefaltRequestContext struct {
	Page *rod.Page
	//TODO: some prox implmentation
}

func New() *DefaltRequestContext {
	return &DefaltRequestContext{}
}

func (r *DefaltRequestContext) Initialise() {

}

func (r *DefaltRequestContext) PerformRequestInstruction(ins RequestInstruction) error {
	r.Page.Navigate(ins.URL)
	r.Page.MustWaitDOMStable()
	return nil
}

func (r *DefaltRequestContext) GetRequestAgent() (interface{}, error) {
	return r.Page, nil
}
