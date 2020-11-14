package mongo

import (
	"context"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

type DefaultSearchResultBuilder struct {
	Database     *mongo.Database
	QueryBuilder QueryBuilder
	SortBuilder  SortBuilder
	Mapper       Mapper
}

func NewSearchResultBuilderWithMapper(db *mongo.Database, queryBuilder QueryBuilder, sortBuilder SortBuilder, mapper Mapper) *DefaultSearchResultBuilder {
	builder := &DefaultSearchResultBuilder{db, queryBuilder, sortBuilder, mapper}
	return builder
}
func NewSearchResultBuilder(db *mongo.Database, queryBuilder QueryBuilder, sortBuilder SortBuilder) *DefaultSearchResultBuilder {
	return NewSearchResultBuilderWithMapper(db, queryBuilder, sortBuilder, nil)
}
func NewMongoSearchResultBuilder(db *mongo.Database, queryBuilder QueryBuilder) *DefaultSearchResultBuilder {
	sortBuilder := &DefaultSortBuilder{}
	return NewSearchResultBuilderWithMapper(db, queryBuilder, sortBuilder, nil)
}

func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, collection *mongo.Collection, m interface{}, modelType reflect.Type) (interface{}, int64, error) {
	query, fields := b.QueryBuilder.BuildQuery(m, modelType)

	var sort = bson.M{}
	var searchModel *search.SearchModel

	if sModel, ok := m.(*search.SearchModel); ok {
		searchModel = sModel
		sort = b.SortBuilder.BuildSort(*sModel, modelType)
	} else {
		value := reflect.Indirect(reflect.ValueOf(m))
		numField := value.NumField()
		for i := 0; i < numField; i++ {
			if sModel1, ok := value.Field(i).Interface().(*search.SearchModel); ok {
				searchModel = sModel1
				sort = b.SortBuilder.BuildSort(*sModel1, modelType)
			}
		}
	}
	return b.build(ctx, collection, modelType, query, fields, sort, searchModel.PageIndex, searchModel.PageSize, searchModel.FirstPageSize)
}

func (b *DefaultSearchResultBuilder) build(ctx context.Context, collection *mongo.Collection, modelType reflect.Type, query bson.M, fields bson.M, sort bson.M, pageIndex int64, pageSize int64, initPageSize int64) (interface{}, int64, error) {
	optionsFind := options.Find()
	optionsFind.Projection = fields
	if initPageSize > 0 {
		if pageIndex == 1 {
			optionsFind.SetSkip(0)
			optionsFind.SetLimit(initPageSize)
		} else {
			optionsFind.SetSkip(pageSize*(pageIndex-2) + initPageSize)
			optionsFind.SetLimit(pageSize)
		}
	} else {
		optionsFind.SetSkip(pageSize * (pageIndex - 1))
		optionsFind.SetLimit(pageSize)
	}
	if sort != nil {
		optionsFind.SetSort(sort)
	}

	databaseQuery, er0 := collection.Find(ctx, query, optionsFind)
	if er0 != nil {
		return nil, 0, er0
	}

	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	results := reflect.New(modelsType).Interface()
	er1 := databaseQuery.All(ctx, results)
	if er1 != nil {
		return results, 0, er1
	}
	options := options.Count()
	count, er2 := collection.CountDocuments(ctx, query, options)
	if er2 != nil {
		return results, 0, er2
	}
	r2, er3 := b.Mapper.DbToModels(ctx, results)
	return r2, count, er3
}
