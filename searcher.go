package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Searcher struct {
	search func(ctx context.Context, m interface{}) (interface{}, int64, error)
}

func NewSearcher(search func(ctx context.Context, m interface{}) (interface{}, int64, error)) *Searcher {
	return &Searcher{search: search}
}

func (s *Searcher) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.search(ctx, m)
}

func NewSearcherWithExtract(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), buildSort func(s string, modelType reflect.Type) bson.M, extract func(m interface{}) (string, int64, int64, int64, error), options...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	builder := NewSearchBuilder(db, collectionName, modelType, buildQuery, buildSort, extract, options...)
	return NewSearcher(builder.Search)
}
func NewSearcherWithMap(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), buildSort func(s string, modelType reflect.Type) bson.M, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (string, int64, int64, int64, error)) *Searcher {
	builder := NewSearchBuilderWithMap(db, collectionName, modelType, buildQuery, buildSort, mp, options...)
	return NewSearcher(builder.Search)
}
func NewSearcherWithQuery(db *mongo.Database, collectionName string, modelType reflect.Type, buildQuery func(sm interface{}) (bson.M, bson.M), options ...func(context.Context, interface{}) (interface{}, error)) *Searcher {
	builder := NewSearcherWithQuery(db, collectionName, modelType, buildQuery, options...)
	return NewSearcher(builder.Search)
}
func NewDefaultSearcher(db *mongo.Database, collectionName string, modelType reflect.Type, mp func(context.Context, interface{}) (interface{}, error), options ...func(m interface{}) (string, int64, int64, int64, error)) *Searcher {
	builder := NewDefaultSearchBuilder(db, collectionName, modelType, mp, options...)
	return NewSearcher(builder.Search)
}
