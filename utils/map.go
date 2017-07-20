package utils

import (
	"reflect"
	"sort"
)

// SortedStringKeys return keys with sorted orders, key must be string
func SortedStringKeys(m interface{}) []string {
	v := reflect.ValueOf(m)
	var keys []string
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keys = append(keys, key.Interface().(string))
		}
		sort.Strings(keys)
	}
	return keys
}
