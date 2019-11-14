package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoGenericSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, idObjectId bool) (*DefaultGenericService, *DefaultSearchService) {
	genericService := NewDefaultGenericService(db, modelType, collectionName, idObjectId)
	searchService := NewDefaultSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}
