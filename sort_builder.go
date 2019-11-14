package mongo

import (
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

type SortBuilder interface {
	BuildSort(searchModel search.SearchModel, modelType reflect.Type) bson.M
}
