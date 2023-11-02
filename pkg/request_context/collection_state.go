package requestcontext

import (
	"github.com/go-rod/rod/lib/proto"
)

const (
	FAILED = iota
	DONE
	PROCESSING
	SKIPPING
)

type CollectionState struct {
	RequestsMade []*proto.NetworkResponseReceived
	LoadState    int
}
