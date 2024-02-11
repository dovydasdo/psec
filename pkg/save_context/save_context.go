package savecontext

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: make ctx handling better in the whole project
type Saver interface {
	Exec(query string, data ...any) (string, error)
	QueryExists(query string, result any, data ...interface{}) (interface{}, error)
}

type PSQLSaver struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

var (
	pgInstance *PSQLSaver
	pgOnce     sync.Once
)

func NewPSQLSaver(ctx context.Context, opts *PSQLOptions) (*PSQLSaver, error) {
	var err error

	pgOnce.Do(func() {
		var pool *pgxpool.Pool
		pool, err = pgxpool.New(ctx, opts.ConString)
		pgInstance = &PSQLSaver{db: pool, logger: opts.Logger}
	})

	return pgInstance, err
}

func (s *PSQLSaver) Exec(query string, data ...any) (string, error) {
	s.logger.Debug("psql", "executing", query, "args", data)
	st, err := s.db.Exec(context.Background(), query, data...)
	if err != nil {
		s.logger.Error("psql", "failed", query, "error", err)
	}
	return st.String(), err
}

func (s *PSQLSaver) QueryExists(query string, result any, data ...interface{}) (interface{}, error) {
	s.logger.Debug("psql", "executing", query, "args", data)
	row := s.db.QueryRow(context.Background(), query, data...)
	err := row.Scan(result)
	if err != nil {
		s.logger.Error("psql", "failed", query, "error", err)
	}
	return result, err
}

// ?
type QueryFailedError struct {
	Message string
	BaseErr error
}

func (e QueryFailedError) Error() string {
	return e.Message
}
