package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

func NewMongoSearchWriterWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), idObjectId bool, versionField string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, versionField, mapper)
	if mapper != nil {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	} else {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	}
}
func NewMongoSearchWriter(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), idObjectId bool, versionField string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, idObjectId, versionField, mapper)
	searcher := NewSearcher(search)
	return writer, searcher
}
func NewSearchWriterWithVersionAndQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), version string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, version, mapper)
	if mapper != nil {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	} else {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	}
}
func NewSearchWriterWithVersion(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), version string, options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, version, mapper)
	searcher := NewSearcher(search)
	return writer, searcher
}
func NewSearchWriterWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, "", mapper)
	if mapper != nil {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery, mapper.DbToModel)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	} else {
		builder := NewSearchBuilder(db.Collection(collectionName), modelType, buildQuery)
		searcher := NewSearcher(builder.Search)
		return writer, searcher
	}
}
func NewSearchWriter(db *mongo.Database, collectionName string, modelType reflect.Type, search func(ctx context.Context, searchModel interface{}) (interface{}, int64, error), options ...Mapper) (*Writer, *Searcher) {
	var mapper Mapper
	if len(options) >= 1 {
		mapper = options[0]
	}
	writer := NewWriterWithVersion(db, collectionName, modelType, false, "", mapper)
	searcher := NewSearcher(search)
	return writer, searcher
}
