package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

type QueryBuilder interface {
	BuildQuery(searchModel interface{}, resultModelType reflect.Type) (bson.M, bson.M)
}
