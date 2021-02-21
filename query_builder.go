package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

type QueryBuilder interface {
	BuildQuery(searchModel interface{}) (bson.M, bson.M)
}
