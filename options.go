package psec

import (
	"log/slog"
)

type Option func(opts *Options)

type Options struct {
	RequestAgentsOpts []interface{}
	SaverOpts         []interface{}
	ProxyAgentOpts    []interface{}
	Logger            *slog.Logger
}

func NewOptions(setters []Option) *Options {
	options := &Options{}

	for _, setter := range setters {
		setter(options)
	}

	return options
}

func WithRequestAgent(agent interface{}) Option {
	return func(opts *Options) {
		opts.RequestAgentsOpts = append(opts.RequestAgentsOpts, agent)
	}
}

func WithSaver(saver interface{}) Option {
	return func(opts *Options) {
		opts.SaverOpts = append(opts.SaverOpts, saver)
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

func WithProxyAgent(agent interface{}) Option {
	return func(opts *Options) {
		opts.ProxyAgentOpts = append(opts.ProxyAgentOpts, agent)
	}
}
