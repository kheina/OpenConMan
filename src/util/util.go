package util

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
)

func OptionalString(str string) *string {
	if str != "" {
		return &str
	}
	return nil
}

// IsNil checks if the interface is nil
func IsNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Func:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// pathExists returns whether or not the given file path exists as a file or
// directory
func PathExists(path string) bool {
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return true
	case errors.Is(err, fs.ErrNotExist):
		return false
	default:
		// only occurs when err != nil, but it's not a "does not exist error"
		// i.e. wtf?
		return true
	}
}
