package psec

import (
	"errors"
	"log"
	"log/slog"

	"github.com/dovydasdo/psec/config"
	r "github.com/dovydasdo/psec/pkg/request_context"
	sc "github.com/dovydasdo/psec/pkg/save_context"
	uc "github.com/dovydasdo/psec/pkg/util_context"
	perrors "github.com/dovydasdo/psec/util/errors"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type ExtractionFunc func(c r.Loader, s sc.Saver) error

type PSEC struct {
	rctx   r.Loader
	sctx   sc.Saver
	uctx   uc.UtilInterface
	cFunc  ExtractionFunc
	cfg    *config.Conf
	logger *slog.Logger
}

// todo: pass config to constructor instead of confiugrating with seperate mothods
func New(l *slog.Logger) *PSEC {
	// temp for testing
	reqConf := config.NewCDPLaunchConf()

	ec := &PSEC{
		rctx:   r.GetCDPContext(*reqConf, l),
		logger: l,
	}

	return ec
}

func (c *PSEC) SetPSQLSaver() *PSEC {
	if c.sctx != nil {
		c.logger.Warn("db", "message", "psql saver called when saver is already initiated")
		return c
	}

	if c.cfg == nil {
		c.logger.Error("db", "message", "config needs to be initialzed before setting a saver")
		return c
	}

	c.sctx = sc.NewPSQLSaver(c.cfg, c.logger)
	return c
}

func (c *PSEC) SetSQLiteSaver() *PSEC {
	if c.sctx != nil {
		c.logger.Warn("db", "message", "sqlite saver called when saver is already initiated")
		return c
	}

	if c.cfg == nil {
		c.logger.Error("db", "message", "config needs to be initialzed before setting a saver")
		return c
	}

	// c.sctx = sc.NewSQLiteSaver()
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
			log.Println("Got nil error, collection complete, terminating")
			return nil
		case perrors.Blocked:
			log.Printf("Got blocked error, resetting and retrying, err: %v", err.Error())
			err = c.rctx.ChangeProxy()
			if err != nil {
				// If no proxies, terminate immediately
				return nil
			}

			c.rctx.Reset()
			continue
		case perrors.ExtractionFailed:
			log.Println("Got extraction error, retrying")
			c.rctx.Reset()
			continue
		default:
			return v
		}
	}

	log.Printf("failed to successfully complete in %v attempts, terminating", limit)

	return nil
}
