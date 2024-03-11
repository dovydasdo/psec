package psec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	r "github.com/dovydasdo/psec/pkg/request_context"
	sc "github.com/dovydasdo/psec/pkg/save_context"
	perrors "github.com/dovydasdo/psec/util/errors"
)

type ExtractionFunc func(c r.Loader, s sc.Saver, l *slog.Logger) error

type PSEC struct {
	rctx   r.Loader
	sctx   sc.Saver
	cFunc  ExtractionFunc
	logger *slog.Logger
}

func New(options *Options) *PSEC {
	ec := &PSEC{
		logger: options.Logger,
	}

	// Set desired request agents
	for _, rao := range options.RequestAgentsOpts {
		switch v := rao.(type) {
		case *r.CDPOptions:
			if ec.rctx != nil {
				// only single agent for now
				break
			}
			ec.rctx = r.GetCDPContext(v)
			break
		default:
			ec.logger.Warn("init", "message", "provided request agent is not supported")
		}
	}

	// Set desired savers
	for _, sao := range options.SaverOpts {
		switch v := sao.(type) {
		case *sc.PSQLOptions:
			if ec.sctx != nil {
				// only single for now
				break
			}
			var err error
			ec.sctx, err = sc.NewPSQLSaver(context.TODO(), v)
			if err != nil {
				ec.logger.Error("psec", "message", "failed to get psql saver", "error", err)
			}
		default:
			ec.logger.Warn("init", "message", "provided saver is not supported")
		}
	}

	// Set desired proxy agents
	for _, pao := range options.ProxyAgentOpts {
		switch v := pao.(type) {
		case *r.BDProxyOptions:
			agent := r.NewBDProxyAgent(v)
			ec.rctx.RegisterProxyAgent(agent)
		default:
			ec.logger.Warn("init", "message", "provided proxy agent is not supported")
		}
	}

	return ec
}

func (c *PSEC) AddSaver(s sc.Saver) error {
	if c.sctx != nil {
		return errors.New("for now only one saver is supported")
	}
	c.sctx = s
	return nil
}

func (c *PSEC) AddRequestAgent(r r.Loader) error {
	if c.rctx != nil {
		return errors.New("for now only one loader is supported")
	}
	c.rctx = r
	return nil
}

func (c *PSEC) InitRequestContext() error {
	return c.rctx.Initialize()
}

func (c *PSEC) AddStartFunc(startFunc ExtractionFunc) {
	c.cFunc = startFunc
}

func (c *PSEC) Start(limit int) error {
	if c.cFunc == nil {
		return errors.New("no stat funcion has been porvided")
	}

	// TODO: allow custom actions from errors
	for i := 0; i < limit; i++ {
		err := c.cFunc(c.rctx, c.sctx, c.logger)

		switch v := err.(type) {
		case nil:
			// Succesfull run should eventually return nil as error
			c.logger.Info("psec", "message", "Got nil error, collection complete, terminating")
			return nil
		case perrors.Blocked:
			c.logger.Info("psec", "message", "Got blocked error, resetting and retrying", "error", err.Error())
			err = c.rctx.ChangeProxy()
			if err != nil {
				// If no proxies, terminate immediately
				return nil
			}

			c.rctx.Reset()
			continue
		case perrors.ExtractionFailed:
			c.logger.Info("psec", "message", "Got extraction error, retrying")
			c.rctx.Reset()
			continue
		default:
			return v
		}
	}

	c.logger.Info("psec", "message", fmt.Sprintf("failed to successfully complete in %v attempts, terminating", limit))

	return nil
}
