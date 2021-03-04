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

func NewMapper(modelType reflect.Type, options ...string) *DefaultMapper {
	var bsonName, latitudeName, longitudeName string
	if len(options) >= 1 && len(options[0]) > 0 {
		bsonName = options[0]
	}
	if len(options) >= 2 && len(options[1]) > 0 {
		latitudeName = options[1]
	} else {
		latitudeName = "Latitude"
	}
	if len(options) >= 3 && len(options[2]) > 0 {
		longitudeName = options[2]
	} else {
		longitudeName = "Longitude"
	}
	latitudeIndex := FindFieldIndex(modelType, latitudeName)
	longitudeIndex := FindFieldIndex(modelType, longitudeName)
	var bsonIndex int
	if len(bsonName) > 0 {
		bsonIndex = FindFieldIndex(modelType, bsonName)
	} else {
		bsonIndex = FindLocationIndex(modelType)
	}

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
	k := valueModelObject.Kind()
	if k == reflect.Map || k == reflect.Struct {
		s.bsonToLocation(valueModelObject, s.bsonIndex, s.latitudeIndex, s.longitudeIndex)
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
			s.bsonToLocation(valueModelObject.Index(i), s.bsonIndex, s.latitudeIndex, s.longitudeIndex)
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
		LocationToBson(valueModelObject, s.bsonIndex, s.latitudeIndex, s.longitudeIndex)
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
			LocationToBson(valueModelObject.Index(i), s.bsonIndex, s.latitudeIndex, s.longitudeIndex)
		}
	}
	return model, nil
}

func BsonToLocation(value reflect.Value, bsonIndex int, latitudeIndex int, longitudeIndex int) {
	if value.Kind() == reflect.Struct {
		x := reflect.Indirect(value)
		b := x.Field(bsonIndex)
		k := b.Kind()
		if k == reflect.Struct || (k == reflect.Ptr && b.IsNil() == false) {
			arrLatLong := reflect.Indirect(b).FieldByName("Coordinates").Interface()
			latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
			longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

			latField := x.Field(latitudeIndex)
			if latField.Kind() == reflect.Ptr {
				var f *float64
				var f2 = latitude.(float64)
				f = &f2
				latField.Set(reflect.ValueOf(f))
			} else {
				latField.Set(reflect.ValueOf(latitude))
			}
			lonField := x.Field(longitudeIndex)
			if lonField.Kind() == reflect.Ptr {
				var f *float64
				var f2 = latitude.(float64)
				f = &f2
				lonField.Set(reflect.ValueOf(f))
			} else {
				lonField.Set(reflect.ValueOf(longitude))
			}
		}
	}
}
func (s *DefaultMapper) bsonToLocation(value reflect.Value, bsonIndex int, latitudeIndex int, longitudeIndex int) {
	if value.Kind() == reflect.Struct {
		x := reflect.Indirect(value)
		b := x.Field(bsonIndex)
		k := b.Kind()
		if k == reflect.Struct || (k == reflect.Ptr && b.IsNil() == false) {
			arrLatLong := reflect.Indirect(b).FieldByName("Coordinates").Interface()
			latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
			longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

			latField := x.Field(latitudeIndex)
			if latField.Kind() == reflect.Ptr {
				var f *float64
				var f2 = latitude.(float64)
				f = &f2
				latField.Set(reflect.ValueOf(f))
			} else {
				latField.Set(reflect.ValueOf(latitude))
			}
			lonField := x.Field(longitudeIndex)
			if lonField.Kind() == reflect.Ptr {
				var f *float64
				var f2 = latitude.(float64)
				f = &f2
				lonField.Set(reflect.ValueOf(f))
			} else {
				lonField.Set(reflect.ValueOf(longitude))
			}
		}
	}

	if value.Kind() == reflect.Map {
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

		arrLatLong := reflect.Indirect(reflect.ValueOf(value.MapIndex(reflect.ValueOf(arrLatLongTag)).Interface())).MapIndex(reflect.ValueOf("coordinates")).Interface()
		latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
		longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

		value.SetMapIndex(reflect.ValueOf(latitudeTag), reflect.ValueOf(latitude))
		value.SetMapIndex(reflect.ValueOf(longitudeTag), reflect.ValueOf(longitude))

		//delete field
		value.SetMapIndex(reflect.ValueOf(arrLatLongTag), reflect.Value{})
	}
}

func LocationToBson(value reflect.Value, bsonIndex int, latitudeIndex int, longitudeIndex int) {
	v := reflect.Indirect(value)
	latitudeField := v.Field(latitudeIndex)
	latNil := false
	if latitudeField.Kind() == reflect.Ptr {
		if latitudeField.IsNil() {
			latNil = true
		}
		latitudeField = reflect.Indirect(latitudeField)
	}

	longNil := false
	longitudeField := v.Field(longitudeIndex)
	if longitudeField.Kind() == reflect.Ptr {
		if longitudeField.IsNil() {
			longNil = true
		}
		longitudeField = reflect.Indirect(longitudeField)
	}
	if latNil == false && longNil == false {
		latitude := latitudeField.Interface()
		longitude := longitudeField.Interface()
		var arr []float64
		arr = append(arr, latitude.(float64), longitude.(float64))

		locationField := v.Field(bsonIndex)
		if locationField.Kind() == reflect.Ptr {
			m := &MongoLocation{Type: "Point", Coordinates: arr}
			locationField.Set(reflect.ValueOf(m))
		} else {
			x := locationField.FieldByName("Type")
			x.Set(reflect.ValueOf("Point"))
			y := locationField.FieldByName("Coordinates")
			y.Set(reflect.ValueOf(arr))
		}
	}
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
