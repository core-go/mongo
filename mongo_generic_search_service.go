package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, idObjectId bool, versionField string, options ...Mapper) (*GenericService, *SearchService) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	genericService := NewMongoGenericService(db, modelType, collectionName, idObjectId, versionField, mapper)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}
func NewGenericSearchServiceWithVersion(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, version string, options ...Mapper) (*GenericService, *SearchService) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	genericService := NewMongoGenericService(db, modelType, collectionName, false, version, mapper)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}
func NewGenericSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, options ...Mapper) (*GenericService, *SearchService) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	genericService := NewMongoGenericService(db, modelType, collectionName, false, "", mapper)
	searchService := NewSearchService(db, modelType, collectionName, searchBuilder)
	return genericService, searchService
}