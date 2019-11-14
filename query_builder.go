package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

type QueryBuilder interface {
	BuildQuery(sm interface{}, resultModelType reflect.Type) bson.M
}
