package savecontext

import (
	"fmt"
	"log"
	"log/slog"
	"reflect"

	"github.com/dovydasdo/psec/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const fmtDBString = "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"

type SaveFunc func(saver Saver, data interface{}) error

type Saver interface {
	Save(data interface{}) error
	Migrate(data interface{}) error
	Exists(id uint, dType interface{}) (interface{}, error)
	CustomSave(saveF SaveFunc, data interface{}) error
}

type PSQLSaver struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewPSQLSaver(cfg *config.Conf, l *slog.Logger) *PSQLSaver {
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
		db:     db,
		logger: l,
	}
}

func (s *PSQLSaver) Save(data interface{}) error {
	//todo: handle migration and make this better. For the record i dont like this but it do be what it do be
	s.db.Create(data)
	return nil
}

func (s *PSQLSaver) CustomSave(saveF SaveFunc, data interface{}) error {
	return saveF(s, data)
}

func (s *PSQLSaver) Exists(id uint, dType interface{}) (interface{}, error) {
	// Cringe
	dataType := reflect.TypeOf(dType).Elem() // Get the type of the element (assuming data is a pointer)

	// Create a new instance of the type
	newData := reflect.New(dataType).Interface()
	r := s.db.Model(dType).Where("id = ?", id).Limit(1).Find(&newData)
	if r.Error != nil {
		return false, &QueryFailedError{
			Message: r.Error.Error(),
			BaseErr: r.Error,
		}
	}

	if r.RowsAffected == 0 {
		return nil, fmt.Errorf("not found")
	}

	return newData, nil
}

func (s *PSQLSaver) Migrate(data interface{}) error {
	t := reflect.TypeOf(data)
	err := s.db.Model(&t).AutoMigrate(data)
	return err
}

type QueryFailedError struct {
	Message string
	BaseErr error
}

func (e QueryFailedError) Error() string {
	return e.Message
}
