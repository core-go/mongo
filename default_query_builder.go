package mongo

import (
	"errors"
	"fmt"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"
	"strings"
)

type DefaultQueryBuilder struct {
	ModelType reflect.Type
}

func NewQueryBuilder(resultModelType reflect.Type) *DefaultQueryBuilder {
	return &DefaultQueryBuilder{ModelType: resultModelType}
}
func (b *DefaultQueryBuilder) BuildQuery(sm interface{}) (bson.M, bson.M) {
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
	}
	for i := 0; i < numField; i++ {
		x := value.Field(i).Interface()
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
		} else if value.Field(i).Kind().String() == "slice" {
			actionDateQuery := bson.M{}
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			actionDateQuery["$in"] = x
			query[columnName] = actionDateQuery
		} else if value.Field(i).Kind().String() == "string" {
			var keywordQuery primitive.Regex
			columnName := GetBsonName(resultModelType, value.Type().Field(i).Name)
			var searchValue string
			if value.Field(i).Len() > 0 {
				const defaultKey = "contain"
				if key, ok := value.Type().Field(i).Tag.Lookup("match"); ok {
					if format, exist := keywordFormat[key]; exist {
						searchValue = fmt.Sprintf(format, value.Field(i).Interface())
					} else {
						log.Panicf("match not support \"%v\" format\n", key)

					}
				} else if format, exist := keywordFormat[defaultKey]; exist {
					searchValue = fmt.Sprintf(format, value.Field(i).Interface())
				}
			} else if len(keyword) > 0 {
				if key, ok := value.Type().Field(i).Tag.Lookup("keyword"); ok {
					if format, exist := keywordFormat[key]; exist {
						searchValue = fmt.Sprintf(format, keyword)

					} else {
						log.Panicf("keyword not support \"%v\" format\n", key)
					}
				}
			}
			if len(searchValue) > 0 {
				keywordQuery = primitive.Regex{Pattern: searchValue}
				query[columnName] = keywordQuery
			}
		} else {
			t := value.Field(i).Kind().String()
			if _, ok := x.(*search.SearchModel); t == "bool" || (strings.Contains(t, "int") && x != 0) || (strings.Contains(t, "float") && x != 0) || (!ok && t == "ptr" &&
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

func ExtractSearchInfo(m interface{}) (string, int64, int64, int64, error) {
	if sModel, ok := m.(*search.SearchModel); ok {
		return sModel.Sort, sModel.Page, sModel.Limit, sModel.FirstLimit, nil
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*search.SearchModel); ok {
				return sModel1.Sort, sModel1.Page, sModel1.Limit, sModel1.FirstLimit, nil
			}
		}
		return "", 0, 0, 0, errors.New("cannot extract sort, pageIndex, pageSize, firstPageSize from model")
	}
}
