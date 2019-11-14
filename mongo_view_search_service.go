package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoViewSearchService(modelType reflect.Type, db *mongo.Database, collection string, searchBuilder SearchResultBuilder, idObjectId bool) (*DefaultViewService, *DefaultSearchService) {
	viewService := NewDefaultViewService(db, modelType, collection, idObjectId)
	searchService := NewDefaultSearchService(db, modelType, collection, searchBuilder)
	return viewService, searchService
}
