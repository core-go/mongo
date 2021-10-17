package query

import (
	"fmt"
	"github.com/core-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"
	"strings"
)

type Builder struct {
	ModelType reflect.Type
}

func NewBuilder(resultModelType reflect.Type) *Builder {
	return &Builder{ModelType: resultModelType}
}
func (b *Builder) BuildQuery(filter interface{}) (bson.M, bson.M) {
	return Build(filter, b.ModelType)
}
func Build(sm interface{}, resultModelType reflect.Type) (bson.M, bson.M) {
	var query = bson.M{}
	var fields = bson.M{}

	if _, ok := sm.(*search.Filter); ok {
		return query, fields
	}

	value := reflect.Indirect(reflect.ValueOf(sm))
	numField := value.NumField()
	var keyword string
	keywordFormat := map[string]string{
		"prefix":  "^%v",
		"contain": "\\w*%v\\w*",
		"equal":   "%v",
	}
	for i := 0; i < numField; i++ {
		field := value.Field(i)
		kind := field.Kind()
		x := field.Interface()
		ps := false
		var psv string
		if kind == reflect.Ptr {
			if field.IsNil() {
				continue
			}
			s0, ok0 := x.(*string)
			if ok0 {
				if s0 == nil || len(*s0) == 0 {
					continue
				}
				ps = true
				psv = *s0
			}
			field = field.Elem()
			kind = field.Kind()
		}
		s0, ok0 := x.(string)
		if ok0 {
			if len(s0) == 0 {
				continue
			}
			psv = s0
		}
		ks := kind.String()
		if v, ok := x.(*search.Filter); ok {
			if len(v.Fields) > 0 {
				for _, key := range v.Fields {
					_, _, columnName := getFieldByJson(resultModelType, key)
					if len(columnName) < 0 {
						fields = bson.M{}
						//fields = fields[len(fields):]
						break
					}
					fields[columnName] = 1
				}
			}
			if v.Excluding != nil && len(v.Excluding) > 0 {
				actionDateQuery := bson.M{}
				actionDateQuery["$nin"] = v.Excluding
			}
			if len(v.Q) > 0 {
				keyword = strings.TrimSpace(v.Q)
			}
			continue
		} else if ps || ks == "string" {
			var keywordQuery primitive.Regex
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			var searchValue string
			var key string
			var ok bool
			if len(psv) > 0 {
				const defaultKey = "contain"
				key, ok = value.Type().Field(i).Tag.Lookup("match")
				if ok {
					if format, exist := keywordFormat[key]; exist {
						searchValue = fmt.Sprintf(format, psv)
					} else {
						log.Panicf("match not support \"%v\" format\n", key)
					}
				} else if format, exist := keywordFormat[defaultKey]; exist {
					searchValue = fmt.Sprintf(format, psv)
				}
			} else if len(keyword) > 0 {
				key, ok = value.Type().Field(i).Tag.Lookup("keyword")
				if ok {
					if format, exist := keywordFormat[key]; exist {
						searchValue = fmt.Sprintf(format, keyword)
					} else {
						log.Panicf("keyword not support \"%v\" format\n", key)
					}
				}
			}
			if len(searchValue) > 0 {
				if key == "equal" {
					query[columnName] = searchValue
				} else {
					keywordQuery = primitive.Regex{Pattern: searchValue}
					query[columnName] = keywordQuery
				}

			}
		} else if rangeTime, ok := x.(*search.TimeRange); ok && rangeTime != nil {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			actionDateQuery["$gte"] = rangeTime.StartTime
			query[columnName] = actionDateQuery
			actionDateQuery["$lt"] = rangeTime.EndTime
			query[columnName] = actionDateQuery
		} else if rangeTime, ok := x.(search.TimeRange); ok {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			actionDateQuery["$gte"] = rangeTime.StartTime
			query[columnName] = actionDateQuery
			actionDateQuery["$lt"] = rangeTime.EndTime
			query[columnName] = actionDateQuery
		} else if rangeDate, ok := x.(*search.DateRange); ok && rangeDate != nil {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			if rangeDate.Min == nil && rangeDate.Max == nil {
				continue
			} else if rangeDate.Max != nil {
				actionDateQuery["$lte"] = rangeDate.Max
			} else if rangeDate.Min != nil {
				actionDateQuery["$gte"] = rangeDate.Min
			} else {
				actionDateQuery["$lte"] = rangeDate.Max
				actionDateQuery["$gte"] = rangeDate.Min
			}
			query[columnName] = actionDateQuery
		} else if rangeDate, ok := x.(search.DateRange); ok {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			if rangeDate.Min == nil && rangeDate.Max == nil {
				continue
			} else if rangeDate.Max != nil {
				actionDateQuery["$lte"] = rangeDate.Max
			} else if rangeDate.Min != nil {
				actionDateQuery["$gte"] = rangeDate.Min
			} else {
				actionDateQuery["$lte"] = rangeDate.Max
				actionDateQuery["$gte"] = rangeDate.Min
			}
			query[columnName] = actionDateQuery
		} else if numberRange, ok := x.(*search.NumberRange); ok && numberRange != nil {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			amountQuery := bson.M{}

			if numberRange.Min != nil {
				amountQuery["$gte"] = *numberRange.Min
			} else if numberRange.Lower != nil {
				amountQuery["$gt"] = *numberRange.Lower
			}
			if numberRange.Max != nil {
				amountQuery["$lte"] = *numberRange.Max
			} else if numberRange.Upper != nil {
				amountQuery["$lt"] = *numberRange.Upper
			}

			if len(amountQuery) > 0 {
				query[columnName] = amountQuery
			}
		} else if numberRange, ok := x.(search.NumberRange); ok {
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			amountQuery := bson.M{}

			if numberRange.Min != nil {
				amountQuery["$gte"] = *numberRange.Min
			} else if numberRange.Lower != nil {
				amountQuery["$gt"] = *numberRange.Lower
			}
			if numberRange.Max != nil {
				amountQuery["$lte"] = *numberRange.Max
			} else if numberRange.Upper != nil {
				amountQuery["$lt"] = *numberRange.Upper
			}

			if len(amountQuery) > 0 {
				query[columnName] = amountQuery
			}
		} else if ks == "slice" {
			actionDateQuery := bson.M{}
			columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery["$in"] = x
			query[columnName] = actionDateQuery
		} else {
			if _, ok := x.(*search.Filter); ks == "bool" || (strings.Contains(ks, "int") && x != 0) || (strings.Contains(ks, "float") && x != 0) || (!ok && ks == "ptr" &&
				value.Field(i).Pointer() != 0) {
				columnName := getBsonName(resultModelType, value.Type().Field(i).Name)
				if len(columnName) > 0 {
					query[columnName] = x
				}
			}
		}
	}
	return query, fields
}

func getFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("json")
		if ok1 && strings.Split(tag1, ",")[0] == jsonName {
			if tag2, ok2 := field.Tag.Lookup("bson"); ok2 {
				return i, field.Name, strings.Split(tag2, ",")[0]
			}
			return i, field.Name, ""
		}
	}
	return -1, jsonName, jsonName
}
func getBsonName(modelType reflect.Type, fieldName string) string {
	field, found := modelType.FieldByName(fieldName)
	if !found {
		return fieldName
	}
	if tag, ok := field.Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return fieldName
}
