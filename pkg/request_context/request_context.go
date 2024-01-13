package requestcontext

type Loader interface {
	RegisterProxyAgent(a ProxyGetter)
	SetBinPath(path string)
	Initialize()
	ChangeProxy() error
	GetState() *State
	ClearState()
	Do(ins ...interface{}) ([]Result, error)
	Reset()
}
