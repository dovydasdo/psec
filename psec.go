package psec

import (
	"errors"
	"fmt"
	"log"

	"github.com/dovydasdo/psec/config"
	r "github.com/dovydasdo/psec/pkg/request_context"
	savecontext "github.com/dovydasdo/psec/pkg/save_context"
	"github.com/dovydasdo/psec/util/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type PSEC struct {
	rctx   r.RequestContextInterface
	sctx   savecontext.Saver
	cFunc  func(c r.RequestContextInterface, s savecontext.Saver) error
	cfg    *config.Conf
	logger logger.Logger
}

func New() *PSEC {
	ec := &PSEC{
		rctx:   r.New(),
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

	c.sctx = savecontext.NewPSQLSaver(db)
	return c
}

func (c *PSEC) AddStartFunc(startFunc func(c r.RequestContextInterface, s savecontext.Saver) error) *PSEC {
	c.cFunc = startFunc
	return c
}

func (c *PSEC) InitEnvConfig() *PSEC {
	c.cfg = config.New()
	return c
}

func (c *PSEC) Start() error {
	if c.cFunc == nil {
		return errors.New("no stat funcion has been porvided")
	}

	return c.cFunc(c.rctx, c.sctx)
}