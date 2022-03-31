package segment

import "github.com/forteilgmbh/segment-config-go/segment"

func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	apiErr, ok := err.(*segment.SegmentApiError)
	if !ok {
		return false
	}
	if apiErr.Code != 404 {
		return false
	}
	return true
}
