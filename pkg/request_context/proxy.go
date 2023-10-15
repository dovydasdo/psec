package requestcontext

type Proxies map[string]Proxy

type Proxy struct {
	Ip         string `json:"ip"`
	TrustScore *int   `json:"trustScore,omitempty"`
}
