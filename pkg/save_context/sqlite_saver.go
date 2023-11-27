package savecontext

import (
	"errors"
	"fmt"
	"log"

	"github.com/dovydasdo/psec/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Only used for turso for now
var dbUrl = "libsql://%s.turso.io?authToken=%s"

type SQLiteSaver struct {
	db *gorm.DB
}

func NewSQLiteSaver() *SQLiteSaver {

	cfg := config.NewTursoConf()

	var logLevel gormlogger.LogLevel
	if cfg.Debug {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	dbString := fmt.Sprintf(dbUrl, cfg.DBName, cfg.DBToken)

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		log.Panic("failed to open db")
	}

	return &SQLiteSaver{
		db: db,
	}
}

func (s *SQLiteSaver) Save(data interface{}) error {
	//todo: handle migration and make this better. For the record i dont like this but it do be what it do be
	s.db.Create(data)
	return nil
}

func (s *SQLiteSaver) Exists(id uint, dType interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (s *SQLiteSaver) Migrate(data interface{}) error {
	err := s.db.AutoMigrate(data)
	return err
}
