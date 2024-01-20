package requestcontext

type BDProxyOption func(opts *BDProxyOptions)

type BDProxyOptions struct {
	Server   string
	Username string
	Password string
}

func NewBDProxyOptions(setters []BDProxyOption) *BDProxyOptions {
	opts := &BDProxyOptions{}

	// TODO: some validation that all of the required options are provided
	for _, setter := range setters {
		setter(opts)
	}

	return opts
}

func WithServer(server string) BDProxyOption {
	return func(opts *BDProxyOptions) {
		opts.Server = server
	}
}
func WithUsername(username string) BDProxyOption {
	return func(opts *BDProxyOptions) {
		opts.Username = username
	}
}
func WithPassword(password string) BDProxyOption {
	return func(opts *BDProxyOptions) {
		opts.Password = password
	}
}
