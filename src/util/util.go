package util

import "reflect"

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
