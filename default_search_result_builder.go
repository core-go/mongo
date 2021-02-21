package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"
)

type DefaultSearchResultBuilder struct {
	Database          *mongo.Database
	QueryBuilder      QueryBuilder
	BuildSort         func(s string, modelType reflect.Type) bson.M
	ExtractSearchInfo func(m interface{}) (string, int64, int64, int64, error)
	Map               func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewSearchResultBuilderWithMapper(db *mongo.Database, queryBuilder QueryBuilder, buildSort func(s string, modelType reflect.Type) bson.M, extractSearchInfo func(m interface{}) (string, int64, int64, int64, error), mp func(context.Context, interface{}) (interface{}, error)) *DefaultSearchResultBuilder {
	builder := &DefaultSearchResultBuilder{Database: db, QueryBuilder: queryBuilder, BuildSort: buildSort, ExtractSearchInfo: extractSearchInfo, Map: mp}
	return builder
}
func NewMongoSearchResultBuilder(db *mongo.Database, queryBuilder QueryBuilder, extractSearchInfo func(m interface{}) (string, int64, int64, int64, error), mp func(context.Context, interface{}) (interface{}, error)) *DefaultSearchResultBuilder {
	return NewSearchResultBuilderWithMapper(db, queryBuilder, BuildSort, extractSearchInfo, mp)
}
func NewSearchResultBuilder(db *mongo.Database, queryBuilder QueryBuilder, extractSearchInfo func(m interface{}) (string, int64, int64, int64, error)) *DefaultSearchResultBuilder {
	return NewSearchResultBuilderWithMapper(db, queryBuilder, BuildSort, extractSearchInfo, nil)
}

func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, collection *mongo.Collection, m interface{}, modelType reflect.Type) (interface{}, int64, error) {
	query, fields := b.QueryBuilder.BuildQuery(m, modelType)

	var sort = bson.M{}
	s, pageIndex, pageSize, firstPageSize, err := b.ExtractSearchInfo(m)
	if err != nil {
		return nil, 0, err
	}
	sort = b.BuildSort(s, modelType)
	return BuildSearchResult(ctx, collection, modelType, query, fields, sort, pageIndex, pageSize, firstPageSize, b.Map)
}
func BuildSearchResult(ctx context.Context, collection *mongo.Collection, modelType reflect.Type, query bson.M, fields bson.M, sort bson.M, pageIndex int64, pageSize int64, initPageSize int64, mp func(context.Context, interface{}) (interface{}, error)) (interface{}, int64, error) {
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
	if mp == nil {
		return results, count, nil
	}
	valueModelObject := reflect.Indirect(reflect.ValueOf(results))
	if valueModelObject.Kind() == reflect.Ptr {
		valueModelObject = reflect.Indirect(valueModelObject)
	}
	if valueModelObject.Kind() == reflect.Slice {
		for i := 0; i < valueModelObject.Len(); i++ {
			_, er3 := mp(ctx, valueModelObject.Index(i))
			if er3 != nil {
				return results, count, er3
			}
		}
	}
	return results, count, nil
}

func BuildSort(s string, modelType reflect.Type) bson.M {
	var sort = bson.M{}
	if len(s) == 0 {
		return sort
	}
	sorts := strings.Split(s, ",")
	for i := 0; i < len(sorts); i++ {
		sortField := strings.TrimSpace(sorts[i])
		fieldName := sortField
		c := sortField[0:1]
		if c == "-" || c == "+" {
			fieldName = sortField[1:]
		}

		columnName := GetBsonNameForSort(modelType, fieldName)
		sortType := GetSortType(c)
		sort[columnName] = sortType
	}
	return sort
}
