package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchLoaderWithQuery(db *mongo.Database, collection string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*Searcher, *Loader) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, collection, modelType, idObjectId, mp)
	builder := NewSearchBuilderWithQuery(db, collection, modelType, buildQuery, mp)
	searcher := NewSearcher(builder.Search)
	return searcher, loader
}
func NewMongoSearchLoader(db *mongo.Database, collection string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*Searcher, *Loader) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	loader := NewMongoLoader(db, collection, modelType, idObjectId, mp)
	searcher := NewSearcher(search)
	return searcher, loader
}

func NewSearchLoader(db *mongo.Database, collection string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), options ...func(context.Context, interface{}) (interface{}, error)) (*Searcher, *Loader) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoSearchLoader(db, collection, modelType, search, false, mp)
}
func NewSearchLoaderWithQuery(db *mongo.Database, collection string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), options ...func(context.Context, interface{}) (interface{}, error)) (*Searcher, *Loader) {
	return NewMongoSearchLoaderWithQuery(db, collection, modelType, buildQuery, false, options...)
}
