package psec

import (
	"errors"
	"fmt"
	"log"

	"github.com/dovydasdo/psec/config"
	r "github.com/dovydasdo/psec/pkg/request_context"
	sc "github.com/dovydasdo/psec/pkg/save_context"
	uc "github.com/dovydasdo/psec/pkg/util_context"
	"github.com/dovydasdo/psec/util/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type PSEC struct {
	rctx   r.RequestContextInterface
	sctx   sc.Saver
	uctx   uc.UtilInterface
	cFunc  func(c r.RequestContextInterface, s sc.Saver, u uc.UtilInterface) error
	cfg    *config.Conf
	logger logger.Logger
}

func New() *PSEC {
	ec := &PSEC{
		rctx:   r.New(),
		uctx:   uc.New(),
		logger: *logger.New(true),
	}

	return ec
}

func (c *PSEC) SetPSQLSaver() *PSEC {
	if c.cfg == nil {
		log.Fatal("config needs to be initialzed before setting a saver")
		return c
	}

	var logLevel gormlogger.LogLevel
	if c.cfg.DB.Debug {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	dbString := fmt.Sprintf(fmtDBString, c.cfg.DB.Host, c.cfg.DB.Username, c.cfg.DB.Password, c.cfg.DB.DBName, c.cfg.DB.Port)

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("DB connection start failure")
	}

	c.sctx = sc.NewPSQLSaver(db)
	return c
}

func (c *PSEC) InitRequestContext() *PSEC {
	c.rctx.Initialize()
	return c
}

func (c *PSEC) AddStartFunc(startFunc func(c r.RequestContextInterface, s sc.Saver, u uc.UtilInterface) error) *PSEC {
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

func (c *PSEC) SetDefaultProxyAgent() *PSEC {
	c.rctx.RegisterProxyAgent(r.NewPSECProxyAgent())
	return c
}

func (c *PSEC) Start() error {
	if c.cFunc == nil {
		return errors.New("no stat funcion has been porvided")
	}

	return c.cFunc(c.rctx, c.sctx, c.uctx)
}
