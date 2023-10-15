package savecontext

import (
	"gorm.io/gorm"
)

type Saver interface {
	Save(data interface{}) error
	Migrate(data interface{}) error
}

type PSQLSaver struct {
	db *gorm.DB
}

func NewPSQLSaver(db *gorm.DB) *PSQLSaver {
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
