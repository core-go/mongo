package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchWriter(db *mongo.Database, modelType reflect.Type, collectionName string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error), idObjectId bool, versionField string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, idObjectId, versionField, mapper)
	searcher := NewSearcher(db, modelType, collectionName, build)
	return writer, searcher
}
func NewSearchWriterWithVersion(db *mongo.Database, modelType reflect.Type, collectionName string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error), version string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, false, version, mapper)
	searcher := NewSearcher(db, modelType, collectionName, build)
	return writer, searcher
}
func NewSearchWriter(db *mongo.Database, modelType reflect.Type, collectionName string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error), options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, modelType, collectionName, false, "", mapper)
	searcher := NewSearcher(db, modelType, collectionName, build)
	return writer, searcher
}
