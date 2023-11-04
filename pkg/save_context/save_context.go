package savecontext

import (
	"fmt"
	"log"

	"github.com/dovydasdo/psec/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type Saver interface {
	Save(data interface{}) error
	Migrate(data interface{}) error
}

type PSQLSaver struct {
	db *gorm.DB
}

func NewPSQLSaver(cfg *config.Conf) *PSQLSaver {
	var logLevel gormlogger.LogLevel
	if cfg.DB.Debug {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	dbString := fmt.Sprintf(fmtDBString, cfg.DB.Host, cfg.DB.Username, cfg.DB.Password, cfg.DB.DBName, cfg.DB.Port)

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		// logger.Fatal().Err(err).Msg("DB connection start failure")
		log.Panic("failed to open db")
	}
	return &PSQLSaver{
		db: db,
	}
}

func (s *PSQLSaver) Save(data interface{}) error {
	//todo: handle migration and make this better. For the record i dont like this but it do be what it do be
	s.db.Create(data)
	return nil
}

func (s *PSQLSaver) Migrate(data interface{}) error {
	err := s.db.AutoMigrate(data)
	return err
}
