package mongo

import (
	"fmt"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"
	"strings"
)

type QueryBuilder struct {
	ModelType reflect.Type
}

func NewQueryBuilder(resultModelType reflect.Type) *QueryBuilder {
	return &QueryBuilder{ModelType: resultModelType}
}
func (b *QueryBuilder) BuildQuery(sm interface{}) (bson.M, bson.M) {
	return BuildQuery(sm, b.ModelType)
}
func BuildQuery(sm interface{}, resultModelType reflect.Type) (bson.M, bson.M) {
	var query = bson.M{}
	var fields = bson.M{}

	if _, ok := sm.(*search.SearchModel); ok {
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
		}
		s0, ok0 := x.(string)
		if ok0 {
			if len(s0) == 0 {
				continue
			}
			psv = s0
		}
		ks := kind.String()
		if v, ok := x.(*search.SearchModel); ok {
			if len(v.Fields) > 0 {
				for _, key := range v.Fields {
					_, _, columnName := GetFieldByJson(resultModelType, key)
					if len(columnName) < 0 {
						fields = bson.M{}
						//fields = fields[len(fields):]
						break
					}
					fields[columnName] = 1
				}
			} else if len(v.Excluding) > 0 {
				for key, val := range v.Excluding {
					idx, fieldName, columnName := GetFieldByJson(resultModelType, key)
					if len(columnName) == 0 {
						if idx >= 0 {
							columnName = fieldName
						} else {
							columnName = key
						}
					}
					if len(val) > 0 {
						actionDateQuery := bson.M{}
						actionDateQuery["$nin"] = val
						query[columnName] = actionDateQuery
					}
				}
			} else if len(v.Keyword) > 0 {
				keyword = strings.TrimSpace(v.Keyword)
			}
			continue
		} else if ps || ks == "string" {
			var keywordQuery primitive.Regex
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
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
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			actionDateQuery["$gte"] = rangeTime.StartTime
			query[columnName] = actionDateQuery
			actionDateQuery["$lt"] = rangeTime.EndTime
			query[columnName] = actionDateQuery
		} else if rangeTime, ok := x.(search.TimeRange); ok {
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			actionDateQuery["$gte"] = rangeTime.StartTime
			query[columnName] = actionDateQuery
			actionDateQuery["$lt"] = rangeTime.EndTime
			query[columnName] = actionDateQuery
		} else if rangeDate, ok := x.(*search.DateRange); ok && rangeDate != nil {
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			if rangeDate.StartDate == nil && rangeDate.EndDate == nil {
				continue
			} else if rangeDate.StartDate == nil {
				actionDateQuery["$lte"] = rangeDate.EndDate
			} else if rangeDate.EndDate == nil {
				actionDateQuery["$gte"] = rangeDate.StartDate
			} else {
				actionDateQuery["$lte"] = rangeDate.EndDate
				actionDateQuery["$gte"] = rangeDate.StartDate
			}
			query[columnName] = actionDateQuery
		} else if rangeDate, ok := x.(search.DateRange); ok {
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery := bson.M{}
			if rangeDate.StartDate == nil && rangeDate.EndDate == nil {
				continue
			} else if rangeDate.StartDate == nil {
				actionDateQuery["$lte"] = rangeDate.EndDate
			} else if rangeDate.EndDate == nil {
				actionDateQuery["$gte"] = rangeDate.StartDate
			} else {
				actionDateQuery["$lte"] = rangeDate.EndDate
				actionDateQuery["$gte"] = rangeDate.StartDate
			}
			query[columnName] = actionDateQuery
		} else if numberRange, ok := x.(*search.NumberRange); ok && numberRange != nil {
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
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
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
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
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery["$in"] = x
			query[columnName] = actionDateQuery
		} else {
			if _, ok := x.(*search.SearchModel); ks == "bool" || (strings.Contains(ks, "int") && x != 0) || (strings.Contains(ks, "float") && x != 0) || (!ok && ks == "ptr" &&
				value.Field(i).Pointer() != 0) {
				columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
				if len(columnName) > 0 {
					query[columnName] = x
				}
			}
		}
	}
	return query, fields
}
