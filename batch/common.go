package batch

import (
	"reflect"
	"strings"
)

func FindIdField(modelType reflect.Type) int {
	return FindField(modelType, "_id")
}
func FindField(modelType reflect.Type, bsonName string) int {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		bsonTag := field.Tag.Get("bson")
		tags := strings.Split(bsonTag, ",")
		for _, tag := range tags {
			if strings.TrimSpace(tag) == bsonName {
				return i
			}
		}
	}
	return -1
}
