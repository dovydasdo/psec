package requestcontext

type Loader interface {
	RegisterProxyAgent(a ProxyGetter)
	SetBinPath(path string)
	Initialize()
	ChangeProxy() error
	GetState() *State
	Do(ins ...interface{}) (string, error)
	Reset()
}
