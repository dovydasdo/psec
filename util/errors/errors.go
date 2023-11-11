package perrors

import "fmt"

// 	Types:
// 		* Blocked
//		* Extraction failed

const (
	BLOCKED_TERMINATE = iota
	BLOCKED_RETRY
)

const (
	EXTRACT_TERMINATE = iota
	EXTRACT_RETRY
)

type Blocked struct {
	SiteId string
	Reason string
	Status int
	Action int
}

func (blocked Blocked) Error() string {
	return fmt.Sprintf("Blocked with status code: %v", blocked.Status)
}

type ExtractionFailed struct {
	SiteId string
	Reason string
	Action int
}

func (f ExtractionFailed) Error() string {
	return fmt.Sprintf("Extraction failed: %v", f.Reason)
}
