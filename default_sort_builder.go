package mongo

import (
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

const desc = "DESC"

type DefaultSortBuilder struct {
}

func (b *DefaultSortBuilder) BuildSort(s search.SearchModel, modelType reflect.Type) bson.M {
	var sort = bson.M{}

	if len(s.SortField) == 0 {
		return sort
	}

	if strings.Index(s.SortField, ",") < 0 {
		columnName := b.getColumnName(s.SortField, modelType)
		sortType := b.getSortType(s.SortType)
		sort[columnName] = sortType
	} else {
		sorts := strings.Split(s.SortField, ",")
		for i := 0; i < len(sorts); i++ {
			sortField := strings.TrimSpace(sorts[i])
			params := strings.Split(sortField, " ")

			if len(params) > 0 {
				columnName := b.getColumnName(params[0], modelType)
				sortType := b.getSortType(params[1])
				sort[columnName] = sortType
			}
		}
	}

	return sort
}

func (b *DefaultSortBuilder) getColumnName(sortField string, modelType reflect.Type) string {
	sortField = strings.TrimSpace(sortField)
	fieldName := b.getFieldNameFromJsonName(sortField, modelType)
	return GetBsonColumnName(modelType, fieldName)
}

func (b *DefaultSortBuilder) getFieldNameFromJsonName(jsonName string, modelType reflect.Type) string {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		if value, ok := field.Tag.Lookup("json"); ok && value == jsonName {
			if value == jsonName {
				return field.Name
			}
		}
	}

	return jsonName
}

func (b *DefaultSortBuilder) getSortType(sortType string) int {
	if strings.ToUpper(sortType) != strings.ToUpper(desc) {
		return 1
	} else {
		return -1
	}
}
