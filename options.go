package psec

import (
	rc "github.com/dovydasdo/psec/pkg/request_context"
	savecontext "github.com/dovydasdo/psec/pkg/save_context"
	"log/slog"
)

type Option func(opts *Options)

type Options struct {
	RequestAgentsOpts []interface{}
	Savers            []interface{}
	Logger            *slog.Logger
}

func NewOptions(setters []Option) *Options {
	options := &Options{}

	for _, setter := range setters {
		setter(options)
	}

	if len(options.RequestAgentsOpts) == 0 {
		cdpOpts := rc.NewCDPOptions(
			[]rc.CDPOption{
				rc.WithInjectionPath("./injection.js"),
			},
		)

		options.Savers = append(options.Savers, cdpOpts)
	}

	return options
}

func WithRequestAgent(agent rc.CDPOptions) Option {
	return func(opts *Options) {
		opts.RequestAgentsOpts = append(opts.RequestAgentsOpts, agent)
	}
}

func WithSaver(saver savecontext.Saver) Option {
	return func(opts *Options) {
		opts.Savers = append(opts.Savers, saver)
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
