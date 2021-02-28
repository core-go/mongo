package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error), idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, modelType, collection, idObjectId, mp)
	searcher := NewSearcher(db, modelType, collection, build)
	return loader, searcher
}

func NewSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error), options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoSearchLoader(db, modelType, collection, build, false, mp)
}
