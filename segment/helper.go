package segment

import (
	"encoding/json"
	"github.com/forteilgmbh/segment-config-go/segment"
	"reflect"
)

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

func IsNilOrZeroValue(v interface{}) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

func toJsonObject(data interface{}) (interface{}, error) {
	val := new(map[string]interface{})
	err := json.Unmarshal([]byte(data.(string)), val)
	return val, err
}

func toJsonArray(data interface{}) (interface{}, error) {
	val := new([]interface{})
	err := json.Unmarshal([]byte(data.(string)), val)
	return val, err
}
