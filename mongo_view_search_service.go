package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, idObjectId bool, options ...Mapper) (*ViewService, *SearchService) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	viewService := NewMongoViewService(db, modelType, collection, idObjectId, mapper)
	searchService := NewSearchService(db, modelType, collection, searchBuilder)
	return viewService, searchService
}

func NewViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, options ...Mapper) (*ViewService, *SearchService) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	return NewMongoViewSearchService(db, modelType, collection, searchBuilder, false, mapper)
}
