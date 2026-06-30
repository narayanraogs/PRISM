package remote

import (
	"sync"
)

var paramsCache sync.Map

// SetParameter sets a single string parameter in the cache.
func SetParameter(param string, value string) {
	paramsCache.Store(param, value)
}

// GetParam retrieves a single string parameter from the cache.
func GetParam(param string) (string, bool) {
	val, ok := paramsCache.Load(param)
	if !ok {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

// SetParameters sets an array of string values for a parameter in the cache.
func SetParameters(param string, values []string) {
	paramsCache.Store(param, values)
}

// GetParams retrieves an array of string values for a parameter from the cache.
func GetParams(param string) ([]string, bool) {
	val, ok := paramsCache.Load(param)
	if !ok {
		return nil, false
	}
	sliceVal, ok := val.([]string)
	return sliceVal, ok
}
