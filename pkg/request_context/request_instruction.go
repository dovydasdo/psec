package requestcontext

import (
	"regexp"

	"github.com/go-rod/rod/lib/proto"
)

const (
	SUCCESS = iota
	TIMEOUT
	BLOCKED
	CONTINUE
	SKIP
)

type RequestInstruction struct {
	URL    string
	Method string
	Filter regexp.Regexp
}

type NavigateInstruction struct {
	URL           string
	DoneCondition interface{}
	Filters       []string
}

type DoneElVisible string

type DoneResponseReceived string

type ReqCtxKey struct {
	Id string
}

type DoneFunc func(resp *proto.NetworkResponse, state *CollectionState) int

type ReqCtxVal struct {
	DoneF        DoneFunc
	RequestsMade []*proto.NetworkResponse
	Source       string
}
