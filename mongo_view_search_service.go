package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) (*MongoLoader, *SearchService) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	viewService := NewMongoLoader(db, modelType, collection, idObjectId, mp)
	searchService := NewSearchService(db, modelType, collection, searchBuilder)
	return viewService, searchService
}

func NewViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, options ...func(context.Context, interface{}) (interface{}, error)) (*MongoLoader, *SearchService) {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoViewSearchService(db, modelType, collection, searchBuilder, false, mp)
}
