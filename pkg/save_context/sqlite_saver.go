package savecontext

import (
	"errors"
)

// Only used for turso for now
var dbUrl = "libsql://%s.turso.io?authToken=%s"

type SQLiteSaver struct {
}

func NewSQLiteSaver() *SQLiteSaver {
	// not implemented yet
	return nil
}

func (s *SQLiteSaver) Save(data interface{}) error {
	//todo: handle migration and make this better. For the record i dont like this but it do be what it do be
	return nil
}

func (s *SQLiteSaver) Exists(id uint, dType interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (s *SQLiteSaver) Migrate(data interface{}) error {
	return nil
}
