package mongo

import (
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
	"time"
)

type DefaultQueryBuilder struct {
}

func (b *DefaultQueryBuilder) BuildQuery(sm interface{}, resultModelType reflect.Type) bson.M {
	var query = bson.M{}

	if _, ok := sm.(*search.SearchModel); ok {
		return query
	}

	value := reflect.Indirect(reflect.ValueOf(sm))
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		if rangeDate, ok := value.Field(i).Interface().(search.DateRange); ok {
			columnName := GetBsonColumnName(resultModelType, value.Type().Field(i).Name)

			actionDateQuery := bson.M{}

			actionDateQuery["$gte"] = rangeDate.StartDate
			query[columnName] = actionDateQuery
			var eDate = rangeDate.EndDate.Add(time.Hour * 24)
			rangeDate.EndDate = &eDate
			actionDateQuery["$lte"] = rangeDate.EndDate
			query[columnName] = actionDateQuery
		} else if rangeTime, ok := value.Field(i).Interface().(search.TimeRange); ok {
			columnName := GetBsonColumnName(resultModelType, value.Type().Field(i).Name)

			actionDateQuery := bson.M{}

			actionDateQuery["$gte"] = rangeTime.StartTime
			query[columnName] = actionDateQuery
			actionDateQuery["$lt"] = rangeTime.EndTime
			query[columnName] = actionDateQuery
		} else {
			if _, ok := value.Field(i).Interface().(*search.SearchModel); value.Field(i).Kind().String() == "bool" || (strings.Contains(value.Field(i).Kind().String(), "int") && value.Field(i).Interface() != 0) || (strings.Contains(value.Field(i).Kind().String(), "float") && value.Field(i).Interface() != 0) || (!ok && value.Field(i).Kind().String() == "string" && value.Field(i).Len() > 0) || (!ok && value.Field(i).Kind().String() == "ptr" &&
				value.Field(i).Pointer() != 0) {
				columnName := GetBsonColumnName(resultModelType, value.Type().Field(i).Name)
				if len(columnName) > 0 {
					query[columnName] = value.Field(i).Interface()
				}
			}
		}
	}
	return query
}
