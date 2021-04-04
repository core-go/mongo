package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchWriterWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), idObjectId bool, versionField string, options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, versionField, mapper)
	if mapper != nil {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	} else {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	}
}
func NewMongoSearchWriter(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), idObjectId bool, versionField string, options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, versionField, mapper)
	searcher := NewSearcher(search)
	return searcher, writer
}
func NewSearchWriterWithVersionAndQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), version string, options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, version, mapper)
	if mapper != nil {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	} else {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	}
}
func NewSearchWriterWithVersion(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), version string, options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, version, mapper)
	searcher := NewSearcher(search)
	return searcher, writer
}
func NewSearchWriterWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, "", mapper)
	if mapper != nil {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	} else {
		builder := NewSearchBuilderWithQuery(db, collectionName, modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return searcher, writer
	}
}
func NewSearchWriter(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), options ...Mapper) (*Searcher, *Writer) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, "", mapper)
	searcher := NewSearcher(search)
	return searcher, writer
}
