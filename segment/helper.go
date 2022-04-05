package segment

import (
	"encoding/json"
	"github.com/forteilgmbh/segment-config-go/segment"
	"reflect"
	"strings"
)

func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	apiErr, ok := err.(*segment.SegmentApiError)
	if !ok {
		return false
	}
	// special case for destination filters: Segment API returns 400 instead of 404
	// and client library further obfuscates this error for some reason
	if apiErr.Code == 3 && strings.Contains(apiErr.Message, "filter does not exist") {
		return true
	}
	if apiErr.Code != 404 {
		return false
	}
	return true
}

func Is500ValidatePermissionsErr(err error) bool {
	// another special case for destination filters: Segment API returns 500 this time instead of 404
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*segment.SegmentApiError); ok {
		if apiErr.Code == 13 && strings.Contains(apiErr.Message, "failed to validate permissions due to an internal error") {
			return true
		}
	}
	return false
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

func strListToInterfaceList(s []string) []interface{} {
	v := make([]interface{}, 0, len(s))
	for _, ss := range s {
		v = append(v, ss)
	}
	return v
}

func allRequiredKeysInMap(m map[string]interface{}, keys ...string) (bool, []string) {
	missingKeys := make([]string, 0)
	for _, k := range keys {
		if v, ok := m[k]; !ok || IsNilOrZeroValue(v) {
			missingKeys = append(missingKeys, k)
		}
	}
	return len(missingKeys) == 0, missingKeys
}

func anyRequiredKeyInMap(m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

func allForbiddenKeysNotInMap(m map[string]interface{}, keys ...string) (bool, []string) {
	extraKeys := make([]string, 0)
	for _, k := range keys {
		if v, ok := m[k]; ok && !IsNilOrZeroValue(v) {
			extraKeys = append(extraKeys, k)
		}
	}
	return len(extraKeys) == 0, extraKeys
}

func Contains(x string, l []string) bool {
	for _, v := range l {
		if v == x {
			return true
		}
	}
	return false
}
