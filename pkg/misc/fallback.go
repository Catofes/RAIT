package misc

import "reflect"

// Fallback returns the given value, or the fallback value if it is empty or of wrong type
func Fallback(value interface{}, fallback interface{}) interface{} {
	if reflect.TypeOf(value) != reflect.TypeOf(fallback) || reflect.ValueOf(value).IsZero() {
		return fallback
	}
	return value
}
