package repository

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	mgo "github.com/core-go/mongo"
)

type SearchRepository[T any, K any, F any] struct {
	*Repository[T, K]
	BuildQuery func(m F) (bson.D, bson.M)
	GetSort    func(m interface{}) string
	BuildSort  func(s string, modelType reflect.Type) bson.D
	ModelType  reflect.Type
}

func NewSearchRepositoryWithSortAndVersion[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, buildSort func(string, reflect.Type) bson.D, idObjectId bool, versionField string, options ...Mapper[T]) *SearchRepository[T, K, F] {
	repo := NewMongoRepositoryWithVersion[T, K](db, collectionName, idObjectId, versionField, options...)
	var t T
	modelType := reflect.TypeOf(t)
	return &SearchRepository[T, K, F]{Repository: repo, BuildSort: buildSort, GetSort: getSort, BuildQuery: buildQuery, ModelType: modelType}
}

func NewSearchRepositoryWithVersion[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, versionField string, options ...Mapper[T]) *SearchRepository[T, K, F] {
	repo := NewMongoRepositoryWithVersion[T, K](db, collectionName, false, versionField, options...)
	var t T
	modelType := reflect.TypeOf(t)
	return &SearchRepository[T, K, F]{Repository: repo, BuildSort: mgo.BuildSort, GetSort: getSort, BuildQuery: buildQuery, ModelType: modelType}
}
func NewSearchRepository[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, options ...Mapper[T]) *SearchRepository[T, K, F] {
	repo := NewMongoRepositoryWithVersion[T, K](db, collectionName, false, "", options...)
	var t T
	modelType := reflect.TypeOf(t)
	return &SearchRepository[T, K, F]{Repository: repo, BuildSort: mgo.BuildSort, GetSort: getSort, BuildQuery: buildQuery, ModelType: modelType}
}
func (b *SearchRepository[T, K, F]) Search(ctx context.Context, m F, limit int64, skip int64) ([]T, int64, error) {
	var objs []T
	query, fields := b.BuildQuery(m)

	var sort = bson.D{}
	s := b.GetSort(m)
	sort = b.BuildSort(s, b.ModelType)
	if skip < 0 {
		skip = 0
	}
	var total int64
	var err error
	total, err = mgo.BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip)
	if b.Mapper != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			b.Mapper.DbToModel(&objs[i])
		}
	}
	return objs, total, err
}
