package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, idObjectId bool, mapper Mapper) (*ViewService, *SearchService) {
	viewService := NewMongoViewService(db, modelType, collection, idObjectId, mapper)
	searchService := NewSearchService(db, modelType, collection, searchBuilder)
	return viewService, searchService
}

func NewViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder) (*ViewService, *SearchService) {
	return NewMongoViewSearchService(db, modelType, collection, searchBuilder, false, nil)
}
