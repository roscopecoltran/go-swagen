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

// Contains 判断obj是否在target中，target支持的类型arrary,slice,map
func Contains(target interface{}, obj interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}
