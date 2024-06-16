package mongo

import "reflect"

// Update
func getValue(model interface{}, index int) (interface{}, error) {
	vo := reflect.Indirect(reflect.ValueOf(model))
	return vo.Field(index).Interface(), nil
}
func findIndex(model interface{}, fieldName string) int {
	modelType := reflect.Indirect(reflect.ValueOf(model))
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		if modelType.Type().Field(i).Name == fieldName {
			return i
		}
	}
	return -1
}
