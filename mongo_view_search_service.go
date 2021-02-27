package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*MongoLoader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, modelType, collection, idObjectId, mp)
	searcher := NewSearcher(db, modelType, collection, searchBuilder)
	return loader, searcher
}

func NewSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, options ...func(context.Context, interface{}) (interface{}, error)) (*MongoLoader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoSearchLoader(db, modelType, collection, searchBuilder, false, mp)
}
