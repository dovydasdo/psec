package psec

import (
	"errors"
	"log"

	"github.com/dovydasdo/psec/config"
	r "github.com/dovydasdo/psec/pkg/request_context"
	sc "github.com/dovydasdo/psec/pkg/save_context"
	uc "github.com/dovydasdo/psec/pkg/util_context"
	"github.com/dovydasdo/psec/util/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type ExtractionFunc func(c r.Loader, s sc.Saver, u uc.UtilInterface) error

type PSEC struct {
	rctx   r.Loader
	sctx   sc.Saver
	uctx   uc.UtilInterface
	cFunc  ExtractionFunc
	cfg    *config.Conf
	logger logger.Logger
}

func New() *PSEC {
	ec := &PSEC{
		rctx:   r.GetCDPContext(),
		uctx:   uc.New(),
		logger: *logger.New(true),
	}

	return ec
}

func (c *PSEC) SetPSQLSaver() *PSEC {
	if c.sctx != nil {
		c.logger.Warn().Str("db", "psql saver called when saver is already initiated")
		return c
	}

	if c.cfg == nil {
		log.Fatal("config needs to be initialzed before setting a saver")
		return c
	}

	c.sctx = sc.NewPSQLSaver(c.cfg)
	return c
}

func (c *PSEC) SetSQLiteSaver() *PSEC {
	if c.sctx != nil {
		c.logger.Warn().Str("db", "sqlite saver called when saver is already initiated")
		return c
	}

	if c.cfg == nil {
		log.Fatal("config needs to be initialzed before setting a saver")
		return c
	}

	c.sctx = sc.NewSQLiteSaver()
	return c
}

func (c *PSEC) InitRequestContext() *PSEC {
	c.rctx.Initialize()
	return c
}

func (c *PSEC) AddStartFunc(startFunc ExtractionFunc) *PSEC {
	c.cFunc = startFunc
	return c
}

func (c *PSEC) InitEnvConfig() *PSEC {
	c.cfg = config.New()
	return c
}

func (c *PSEC) SetBinPath(path string) *PSEC {
	c.rctx.SetBinPath(path)
	return c
}

func (c *PSEC) RegisterProxyAgent(p r.ProxyGetter) *PSEC {
	c.rctx.RegisterProxyAgent(p)
	return c
}

func (c *PSEC) SetDefaultProxyAgent() *PSEC {
	c.rctx.RegisterProxyAgent(r.NewPSECProxyAgent())
	return c
}

func (c *PSEC) Start() error {
	if c.cFunc == nil {
		return errors.New("no stat funcion has been porvided")
	}

	err := c.cFunc(c.rctx, c.sctx, c.uctx)

	return err
}
