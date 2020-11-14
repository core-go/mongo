package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, idObjectId bool, versionField string, mapper Mapper) (*GenericService, *SearchService) {
	genericService := NewMongoGenericService(db, modelType, collectionName, idObjectId, versionField, mapper)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}
func NewGenericSearchServiceWithVersion(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, version string) (*GenericService, *SearchService) {
	genericService := NewMongoGenericService(db, modelType, collectionName, false, version, nil)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}
func NewGenericSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder) (*GenericService, *SearchService) {
	genericService := NewMongoGenericService(db, modelType, collectionName, false, "", nil)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}