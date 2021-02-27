package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchWriter(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, idObjectId bool, versionField string, options ...Mapper) (*MongoWriter, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, idObjectId, versionField, mapper)
	searcher := NewSearcher(db, modelType, collectionName, searchBuilder)
	return writer, searcher
}
func NewSearchWriterWithVersion(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, version string, options ...Mapper) (*MongoWriter, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, false, version, mapper)
	searcher := NewSearcher(db, modelType, collectionName, searchBuilder)
	return writer, searcher
}
func NewSearchWriter(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder, options ...Mapper) (*MongoWriter, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, false, "", mapper)
	searcher := NewSearcher(db, modelType, collectionName, searchBuilder)
	return writer, searcher
}
