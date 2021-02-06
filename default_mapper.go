package mongo

import (
	"context"
	"reflect"
	"strings"
)

type DefaultMapper struct {
	modelType      reflect.Type
	latitudeIndex  int
	longitudeIndex int
	bsonIndex      int
	latitudeName   string
	longitudeName  string
	bsonName       string
}

func NewMapper(modelType reflect.Type, bsonName string, options ...string) *DefaultMapper {
	var latitudeName, longitudeName string
	if len(options) >= 1 && len(options[0]) > 0 {
		latitudeName = options[0]
	} else {
		latitudeName = "latitude"
	}
	if len(options) >= 2 && len(options[1]) > 0 {
		longitudeName = options[1]
	} else {
		longitudeName = "longitude"
	}
	latitudeIndex := FindFieldIndex(modelType, latitudeName)
	longitudeIndex := FindFieldIndex(modelType, longitudeName)
	bsonIndex := FindFieldIndex(modelType, bsonName)
	return &DefaultMapper{
		modelType:      modelType,
		latitudeIndex:  latitudeIndex,
		longitudeIndex: longitudeIndex,
		bsonIndex:      bsonIndex,
		latitudeName:   latitudeName,
		longitudeName:  longitudeName,
		bsonName:       bsonName,
	}
}

func (s *DefaultMapper) DbToModel(ctx context.Context, model interface{}) (interface{}, error) {
	valueModelObject := reflect.Indirect(reflect.ValueOf(model))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}

	if valueModelObject.Kind() == reflect.Map || valueModelObject.Kind() == reflect.Struct {
		s.doBsonToLocation(valueModelObject)
	}
	return model, nil
}

func (s *DefaultMapper) DbToModels(ctx context.Context, model interface{}) (interface{}, error) {
	valueModelObject := reflect.Indirect(reflect.ValueOf(model))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}

	if valueModelObject.Kind() == reflect.Slice {
		for i := 0; i < valueModelObject.Len(); i++ {
			s.doBsonToLocation(valueModelObject.Index(i))
		}
	}
	return model, nil
}

func (s *DefaultMapper) ModelToDb(ctx context.Context, model interface{}) (interface{}, error) {
	valueModelObject := reflect.Indirect(reflect.ValueOf(model))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}

	if valueModelObject.Kind() == reflect.Struct {
		s.doLocationToBson(valueModelObject)
	}
	return model, nil
}

func (s *DefaultMapper) ModelsToDb(ctx context.Context, model interface{}) (interface{}, error) {
	valueModelObject := reflect.Indirect(reflect.ValueOf(model))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}

	if valueModelObject.Kind() == reflect.Slice {
		for i := 0; i < valueModelObject.Len(); i++ {
			s.doLocationToBson(valueModelObject.Index(i))
		}
	}
	return model, nil
}

func (s *DefaultMapper) doBsonToLocation(value reflect.Value) {
	var arrLatLongTag, latitudeTag, longitudeTag string
	if arrLatLongTag = GetBsonName(s.modelType, s.bsonName); arrLatLongTag == "" || arrLatLongTag == "-" {
		arrLatLongTag = getJsonName(s.modelType, s.bsonName)
	}

	if latitudeTag = GetBsonName(s.modelType, s.latitudeName); latitudeTag == "" || latitudeTag == "-" {
		latitudeTag = getJsonName(s.modelType, s.latitudeName)
	}

	if longitudeTag = GetBsonName(s.modelType, s.longitudeName); longitudeTag == "" || longitudeTag == "-" {
		longitudeTag = getJsonName(s.modelType, s.longitudeName)
	}

	if value.Kind() == reflect.Struct {
		arrLatLong := reflect.Indirect(reflect.Indirect(value).Field(s.bsonIndex)).FieldByName("Coordinates").Interface()
		latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
		longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

		reflect.Indirect(value).Field(s.latitudeIndex).Set(reflect.ValueOf(latitude))
		reflect.Indirect(value).Field(s.longitudeIndex).Set(reflect.ValueOf(longitude))
	}

	if value.Kind() == reflect.Map {
		arrLatLong := reflect.Indirect(reflect.ValueOf(value.MapIndex(reflect.ValueOf(arrLatLongTag)).Interface())).MapIndex(reflect.ValueOf("coordinates")).Interface()
		latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
		longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

		value.SetMapIndex(reflect.ValueOf(latitudeTag), reflect.ValueOf(latitude))
		value.SetMapIndex(reflect.ValueOf(longitudeTag), reflect.ValueOf(longitude))

		//delete field
		value.SetMapIndex(reflect.ValueOf(arrLatLongTag), reflect.Value{})
	}
}

func (s *DefaultMapper) doLocationToBson(value reflect.Value) {
	latitudeField := reflect.Indirect(value).Field(s.latitudeIndex)
	if latitudeField.Kind() == reflect.Ptr {
		latitudeField = reflect.Indirect(latitudeField)
	}

	longitudeField := reflect.Indirect(value).Field(s.longitudeIndex)
	if longitudeField.Kind() == reflect.Ptr {
		longitudeField = reflect.Indirect(longitudeField)
	}

	locationField := reflect.Indirect(value).Field(s.bsonIndex)
	if locationField.Kind() == reflect.Ptr {
		locationField = reflect.Indirect(locationField)
	}

	latitude := latitudeField.Interface()
	longitude := longitudeField.Interface()
	locationField.FieldByName("Type").Set(reflect.ValueOf("Point"))

	var arr []float64
	arr = append(arr, latitude.(float64), longitude.(float64))
	locationField.FieldByName("Coordinates").Set(reflect.ValueOf(arr))
}

func getJsonName(modelType reflect.Type, fieldName string) string {
	field, found := modelType.FieldByName(fieldName)
	if !found {
		return fieldName
	}
	if tag, ok := field.Tag.Lookup("json"); ok {
		return strings.Split(tag, ",")[0]
	}
	return fieldName
}
