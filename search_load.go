package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchLoaderWithQuery(db *mongo.Database, modelType reflect.Type, collection string, buildQuery func(sm interface{}) (bson.M, bson.M), idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, modelType, collection, idObjectId, mp)
	builder := NewSearchBuilder(db.Collection(collection), modelType, buildQuery, mp)
	searcher := NewSearcher(builder.Search)
	return loader, searcher
}
func NewMongoSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, modelType, collection, idObjectId, mp)
	searcher := NewSearcher(search)
	return loader, searcher
}

func NewSearchLoader(db *mongo.Database, modelType reflect.Type, collection string, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoSearchLoader(db, modelType, collection, search, false, mp)
}
func NewSearchLoaderWithQuery(db *mongo.Database, modelType reflect.Type, collection string, buildQuery func(sm interface{}) (bson.M, bson.M), options ...func(context.Context, interface{}) (interface{}, error)) (*Loader, *Searcher) {
	return NewMongoSearchLoaderWithQuery(db, modelType, collection, buildQuery, false, options...)
}
