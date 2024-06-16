package mongo

import (
	"reflect"
	"strings"
)

func MakeBsonMap(modelType reflect.Type) map[string]string {
	maps := make(map[string]string)
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		key1 := field.Name
		if tag0, ok0 := field.Tag.Lookup("json"); ok0 {
			if strings.Contains(tag0, ",") {
				a := strings.Split(tag0, ",")
				key1 = a[0]
			} else {
				key1 = tag0
			}
		}
		if tag, ok := field.Tag.Lookup("bson"); ok {
			if tag != "-" {
				if strings.Contains(tag, ",") {
					a := strings.Split(tag, ",")
					if key1 == "-" {
						key1 = a[0]
					}
					maps[key1] = a[0]
				} else {
					if key1 == "-" {
						key1 = tag
					}
					maps[key1] = tag
				}
			}
		} else {
			if key1 == "-" {
				key1 = field.Name
			}
			maps[key1] = key1
		}
	}
	return maps
}

// For Patch
func MapToBson(object map[string]interface{}, objectMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range object {
		field, ok := objectMap[key]
		if ok {
			result[field] = value
		}
	}
	return result
}
