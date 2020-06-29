package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewViewSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder, idObjectId bool, mapper Mapper) (*ViewService, *SearchService) {
	viewService := NewViewService(db, modelType, collection, idObjectId, mapper)
	searchService := NewSearchService(db, modelType, collection, searchBuilder)
	return viewService, searchService
}

func NewDefaultSearchService(db *mongo.Database, modelType reflect.Type, collection string, searchBuilder SearchResultBuilder) (*ViewService, *SearchService) {
	return NewViewSearchService(db, modelType, collection, searchBuilder, false, nil)
}
