package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewGenericSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, idObjectId bool, versionField string, mapper Mapper) (*GenericService, *SearchService) {
	genericService := NewGenericService(db, modelType, collectionName, idObjectId, versionField, mapper)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}

func NewDefaultGenericSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder) (*GenericService, *SearchService) {
	genericService := NewGenericService(db, modelType, collectionName, false, "", nil)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}