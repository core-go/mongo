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
}

func NewSearchRepositoryWithSortAndVersion[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, buildSort func(string, reflect.Type) bson.D, modelType reflect.Type, idObjectId bool, versionField string, options ...mgo.Mapper) *SearchRepository[T, K, F] {
	var repository *Repository[T, K]
	repository = NewMongoRepositoryWithVersion[T, K](db, collectionName, idObjectId, versionField, options...)
	return &SearchRepository[T, K, F]{Repository: repository, BuildSort: buildSort, GetSort: getSort, BuildQuery: buildQuery}
}

func NewSearchRepositoryWithVersion[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, idObjectId bool, versionField string, options ...mgo.Mapper) *SearchRepository[T, K, F] {
	var adapter *Repository[T, K]
	adapter = NewMongoRepositoryWithVersion[T, K](db, collectionName, idObjectId, versionField, options...)
	return &SearchRepository[T, K, F]{Repository: adapter, BuildSort: mgo.BuildSort, GetSort: getSort, BuildQuery: buildQuery}
}
func NewSearchRepository[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, options ...mgo.Mapper) *SearchRepository[T, K, F] {
	var adapter *Repository[T, K]
	adapter = NewMongoRepositoryWithVersion[T, K](db, collectionName, false, "", options...)
	return &SearchRepository[T, K, F]{Repository: adapter, BuildSort: mgo.BuildSort, GetSort: getSort, BuildQuery: buildQuery}
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
	if b.Mapper != nil {
		total, err = mgo.BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip, b.Mapper.DbToModel)
	} else {
		total, err = mgo.BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip)
	}
	return objs, total, err
}
