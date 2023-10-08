package psec

import (
	"errors"

	"github.com/dovydasdo/psec/config"
	r "github.com/dovydasdo/psec/pkg/request_context"
)

type PSEC struct {
	rctx  r.RequestContextInterface
	cFunc func(c r.RequestContextInterface) error
	cfg   *config.Conf
}

func New() *PSEC {
	ec := &PSEC{
		rctx: r.New(),
	}

	ec.initEnvConfig()
	//Init save context here

	return ec
}

func (c *PSEC) AddStartFunc(startFunc func(c r.RequestContextInterface) error) *PSEC {
	c.cFunc = startFunc
	return c
}

func (c *PSEC) initEnvConfig() *PSEC {
	c.cfg = config.New()
	return c
}

func (c *PSEC) Start() error {
	if c.cFunc == nil {
		return errors.New("no stat funcion has been porvided")
	}

	return c.cFunc(c.rctx)
}
