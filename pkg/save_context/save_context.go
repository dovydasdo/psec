package savecontext

type Saver[T any] interface {
	Save(data T) error
}

func NewPSQLSaver() {

}

type PSQLSaver struct {
}
