package mongo

import (
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

type DefaultSortBuilder struct {
}

func (b *DefaultSortBuilder) BuildSort(s search.SearchModel, modelType reflect.Type) bson.M {
	var sort = bson.M{}
	if len(s.Sort) == 0 {
		return sort
	}
	sorts := strings.Split(s.Sort, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}

		columnName := b.getColumnName(modelType, fieldName)
		sortType := b.getSortType(c)
		sort[columnName] = sortType
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
	if sortType == "-" {
		return -1
	} else  {
		return 1
	}
}
