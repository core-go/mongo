package geo

import (
	"reflect"
	"strings"
)

type PointMapper[T any] struct {
	modelType      reflect.Type
	latitudeIndex  int
	longitudeIndex int
	bsonIndex      int
	bsonName       string
	latitudeJson   string
	longitudeJson  string
}

// For Get By Id
func findFieldIndex(modelType reflect.Type, fieldName string) int {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		if field.Name == fieldName {
			return i
		}
	}
	return -1
}
func getJsonByIndex(modelType reflect.Type, fieldIndex int) string {
	if tag, ok := modelType.Field(fieldIndex).Tag.Lookup("json"); ok {
		return strings.Split(tag, ",")[0]
	}
	return ""
}
func getBsonNameByIndex(modelType reflect.Type, fieldIndex int) string {
	if tag, ok := modelType.Field(fieldIndex).Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return ""
}
func FindGeoIndex(modelType reflect.Type) int {
	numField := modelType.NumField()
	k := JSON{}
	for i := 0; i < numField; i++ {
		t := modelType.Field(i).Type
		if t == reflect.TypeOf(&k) || t == reflect.TypeOf(k) {
			return i
		}
	}
	return -1
}

func NewMapper[T any](options ...string) *PointMapper[T] {
	var t T
	modelType0 := reflect.TypeOf(t)
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
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
	latitudeIndex := findFieldIndex(modelType, latitudeName)
	longitudeIndex := findFieldIndex(modelType, longitudeName)
	var bsonIndex int
	if len(bsonName) > 0 {
		bsonIndex = findFieldIndex(modelType, bsonName)
	} else {
		bsonIndex = FindGeoIndex(modelType)
	}
	bs := getBsonNameByIndex(modelType, bsonIndex)
	latitudeJson := getJsonByIndex(modelType, latitudeIndex)
	longitudeJson := getJsonByIndex(modelType, longitudeIndex)
	return &PointMapper[T]{
		modelType:      modelType0,
		latitudeIndex:  latitudeIndex,
		longitudeIndex: longitudeIndex,
		bsonIndex:      bsonIndex,
		bsonName:       bs,
		latitudeJson:   latitudeJson,
		longitudeJson:  longitudeJson,
	}
}

func (s *PointMapper[T]) DbToModel(model T) T {
	var rv reflect.Value
	if s.modelType.Kind() == reflect.Ptr {
		rv = reflect.Indirect(reflect.ValueOf(model))
	} else {
		rv = reflect.ValueOf(&model).Elem()
	}
	b := rv.Field(s.bsonIndex)
	k := b.Kind()
	if k == reflect.Struct || (k == reflect.Ptr && !b.IsNil()) {
		arrLatLong := reflect.Indirect(b).FieldByName("Coordinates").Interface()
		latitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(0).Interface()
		longitude := reflect.Indirect(reflect.ValueOf(arrLatLong)).Index(1).Interface()

		latField := rv.Field(s.latitudeIndex)
		if latField.Kind() == reflect.Ptr {
			var f2 = latitude.(float64)
			latField.Set(reflect.ValueOf(&f2))
		} else {
			latField.Set(reflect.ValueOf(latitude))
		}
		lonField := rv.Field(s.longitudeIndex)
		if lonField.Kind() == reflect.Ptr {
			var f *float64
			var f2 = longitude.(float64)
			f = &f2
			lonField.Set(reflect.ValueOf(f))
		} else {
			lonField.Set(reflect.ValueOf(longitude))
		}
	}
	return model
}
func (s *PointMapper[T]) MapToDb(m map[string]interface{}) map[string]interface{} {
	return FromPointMap(m, s.bsonName, s.latitudeJson, s.longitudeJson)
}
func (s *PointMapper[T]) ModelToDb(model T) T {
	var rv reflect.Value
	if s.modelType.Kind() == reflect.Ptr {
		rv = reflect.Indirect(reflect.ValueOf(model))
	} else {
		rv = reflect.ValueOf(&model).Elem()
	}
	latitudeField := rv.Field(s.latitudeIndex)
	latNil := false
	if latitudeField.Kind() == reflect.Ptr {
		if latitudeField.IsNil() {
			latNil = true
		}
		latitudeField = reflect.Indirect(latitudeField)
	}
	longNil := false
	longitudeField := rv.Field(s.longitudeIndex)
	if longitudeField.Kind() == reflect.Ptr {
		if longitudeField.IsNil() {
			longNil = true
		}
		longitudeField = reflect.Indirect(longitudeField)
	}
	if !latNil && !longNil {
		latitude := latitudeField.Interface()
		longitude := longitudeField.Interface()
		la, ok3 := latitude.(float64)
		lo, ok4 := longitude.(float64)
		if ok3 && ok4 {
			var arr []float64
			arr = append(arr, la, lo)
			coordinatesField := rv.Field(s.bsonIndex)
			if coordinatesField.Kind() == reflect.Ptr {
				m := &JSON{Type: "Point", Coordinates: arr}
				coordinatesField.Set(reflect.ValueOf(m))
			} else {
				x := coordinatesField.FieldByName("Type")
				x.Set(reflect.ValueOf("Point"))
				y := coordinatesField.FieldByName("Coordinates")
				y.Set(reflect.ValueOf(arr))
			}
		}
	}
	return model
}

func FromPointMap(m map[string]interface{}, bsonName string, latitudeJson string, longitudeJson string) map[string]interface{} {
	latV, ok1 := m[latitudeJson]
	logV, ok2 := m[longitudeJson]
	if ok1 && ok2 && len(bsonName) > 0 {
		la, ok3 := latV.(float64)
		lo, ok4 := logV.(float64)
		if ok3 && ok4 {
			var arr []float64
			arr = append(arr, la, lo)
			ml := JSON{Type: "Point", Coordinates: arr}
			m2 := make(map[string]interface{})
			m2[bsonName] = ml
			for key := range m {
				if key != latitudeJson && key != longitudeJson {
					m2[key] = m[key]
				}
			}
			return m2
		}
	}
	return m
}
