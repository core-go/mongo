package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type SearchResultBuilder interface {
	BuildSearchResult(ctx context.Context, collection *mongo.Collection, searchModel interface{}, modelType reflect.Type) (interface{}, int64, error)
}
