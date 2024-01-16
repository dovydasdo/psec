package psec

import (
	requestcontext "github.com/dovydasdo/psec/pkg/request_context"
	savecontext "github.com/dovydasdo/psec/pkg/save_context"
	"golang.org/x/exp/slog"
)

type Option func(opts *Options)

type Options struct {
	RequestAgents []requestcontext.Loader
	Savers        []savecontext.Saver
	Logger        slog.Logger
}

func NewOptions(setters []Option) *Options {

	options := &Options{
		// Defaults
	}

	for _, setter := range setters {
		setter(options)
	}

	return options
}

func WithRequestAgent(agent requestcontext.Loader) Option {
	return func(opts *Options) {
		opts.RequestAgents = append(opts.RequestAgents, agent)
	}
}

func WithSaver(saver savecontext.Saver) Option {
	return func(opts *Options) {
		opts.Savers = append(opts.Savers, saver)
	}
}

func WithLogger(logger slog.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
