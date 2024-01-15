package requestcontext

import (
	"regexp"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

const (
	SUCCESS = iota
	TIMEOUT
	BLOCKED
	CONTINUE
	SKIP
)

type BaseInstruction struct {
	Name string
}

type RequestInstruction struct {
	*BaseInstruction
	URL    string
	Method string
	Filter regexp.Regexp
}

type NavigateInstruction struct {
	*BaseInstruction
	URL           string
	DoneCondition interface{}
	Filters       []string
}

type JSEvalInstruction struct {
	*BaseInstruction
	Script  string
	Timeout time.Duration
	Result  interface{}
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
