package requestcontext

import "log/slog"

type CDPOption func(*CDPOptions)

type CDPOptions struct {
	BinPath       string
	InjectionPath string
	Logger        *slog.Logger
}

func NewCDPOptions(setters []CDPOption) *CDPOptions {
	options := &CDPOptions{
		// Defualts
		BinPath:       "",
		InjectionPath: "./injection.js",
	}

	for _, setter := range setters {
		setter(options)
	}

	return options
}

func WithBinPath(path string) CDPOption {
	return func(c *CDPOptions) {
		c.BinPath = path
	}
}

func WithInjectionPath(path string) CDPOption {
	return func(c *CDPOptions) {
		c.InjectionPath = path
	}
}

func WithLogger(logger *slog.Logger) CDPOption {
	return func(c *CDPOptions) {
		c.Logger = logger
	}
}
