package savecontext

import "log/slog"

type PSQLOption func(opts *PSQLOptions)

type PSQLOptions struct {
	ConString string
	Logger    slog.Logger
}

func NewPSQLOptions(setters []PSQLOption) *PSQLOptions {
	opts := &PSQLOptions{}

	for _, setter := range setters {
		setter(opts)
	}

	return opts
}

func WithLogger(logger slog.Logger) PSQLOption {
	return func(opts *PSQLOptions) {
		opts.Logger = logger
	}
}

func WithConnString(cstrn string) PSQLOption {
	return func(opts *PSQLOptions) {
		opts.ConString = cstrn
	}
}
