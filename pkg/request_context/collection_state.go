package requestcontext

import "github.com/go-rod/rod/lib/proto"

type CollectionState struct {
	RequestsMade []*proto.NetworkResponse
}
