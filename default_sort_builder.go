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
		columnName := b.getColumnName(modelType, s.SortField)
		sortType := b.getSortType(s.SortType)
		sort[columnName] = sortType
	} else {
		sorts := strings.Split(s.SortField, ",")
		for i := 0; i < len(sorts); i++ {
			sortField := strings.TrimSpace(sorts[i])
			params := strings.Split(sortField, " ")

			if len(params) > 0 {
				columnName := b.getColumnName(modelType, params[0])
				sortType := b.getSortType(params[1])
				sort[columnName] = sortType
			}
		}
	}
	return sort
}

func (b *DefaultSortBuilder) getColumnName(modelType reflect.Type, sortField string) string {
	sortField = strings.TrimSpace(sortField)
	idx, fieldName, name  := GetFieldByJson(modelType, sortField)
	if len(name) > 0 {
		return name
	}
	if idx >= 0 {
		return fieldName
	}
	return sortField
}

func (b *DefaultSortBuilder) getSortType(sortType string) int {
	if strings.ToUpper(sortType) != strings.ToUpper(desc) {
		return 1
	}
	return -1
}
