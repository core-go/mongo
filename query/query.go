package query

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type Query[T any, K any, F any] struct {
	*Loader[T, K]
	BuildQuery func(m F) (bson.D, bson.M)
	GetSort    func(m interface{}) string
	BuildSort  func(s string, modelType reflect.Type) bson.D
}

func NewQuery[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, options ...func(context.Context, interface{}) (interface{}, error)) *Query[T, K, F] {
	var loader *Loader[T, K]
	loader = NewMongoLoader[T, K](db, collectionName, false, options...)
	return &Query[T, K, F]{Loader: loader, BuildSort: mgo.BuildSort, GetSort: getSort, BuildQuery: buildQuery}
}
func NewMongoQuery[T any, K any, F any](db *mongo.Database, collectionName string, buildQuery func(m F) (bson.D, bson.M), getSort func(interface{}) string, buildSort func(string, reflect.Type) bson.D, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) *Query[T, K, F] {
	var loader *Loader[T, K]
	loader = NewMongoLoader[T, K](db, collectionName, idObjectId, options...)
	return &Query[T, K, F]{Loader: loader, BuildSort: buildSort, GetSort: getSort, BuildQuery: buildQuery}
}

func (b *Query[T, K, F]) Search(ctx context.Context, m F, limit int64, skip int64) ([]T, int64, error) {
	var objs []T
	query, fields := b.BuildQuery(m)

	var sort = bson.D{}
	s := b.GetSort(m)
	sort = b.BuildSort(s, b.ModelType)
	if skip < 0 {
		skip = 0
	}
	if b.Map != nil {
		total, err := mgo.BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip, b.Map)
		return objs, total, err
	} else {
		total, err := mgo.BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip)
		return objs, total, err
	}
}
