package requestcontext

import "sync"

type State struct {
	NetworkEvents sync.Map
	Source        string
}

type NetworkEvent struct {
	Request  NetworkRequest
	Response NetworkResponse
	// TODO: timings and stuff maybe
}

type NetworkRequest struct {
	URL     string
	Body    string
	Headers map[string]string
}

type NetworkResponse struct {
	URL     string
	Body    string
	Headers map[string]string
}
