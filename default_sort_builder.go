package mongo

import (
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

type DefaultSortBuilder struct {
}
func NewSortBuilder() *DefaultSortBuilder {
	return &DefaultSortBuilder{}
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

		columnName := GetBsonNameForSort(modelType, fieldName)
		sortType := GetSortType(c)
		sort[columnName] = sortType
	}
	return sort
}
