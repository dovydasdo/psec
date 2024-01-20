package psec

import (
	"errors"
	"log/slog"

	r "github.com/dovydasdo/psec/pkg/request_context"
	sc "github.com/dovydasdo/psec/pkg/save_context"
	uc "github.com/dovydasdo/psec/pkg/util_context"
	perrors "github.com/dovydasdo/psec/util/errors"
)

type ExtractionFunc func(c r.Loader, s sc.Saver) error

type PSEC struct {
	rctx   r.Loader
	sctx   sc.Saver
	uctx   uc.UtilInterface
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
			ec.sctx = sc.NewPSQLSaver(v)
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
		err := c.cFunc(c.rctx, c.sctx)

		switch v := err.(type) {
		case nil:
			// Succesfull run should eventually return nil as error
			c.logger.Info("psec", "message", "Got nil error, collection complete, terminating")
			return nil
		case perrors.Blocked:
			c.logger.Info("psec", "message", "Got blocked error, resetting and retrying, err: %v", err.Error())
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

	c.logger.Info("psec", "message", "failed to successfully complete in %v attempts, terminating", limit)

	return nil
}
