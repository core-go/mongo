package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

func BuildSearchResult(ctx context.Context, collection *mongo.Collection, modelType reflect.Type, query bson.M, fields bson.M, sort bson.M, pageIndex int64, pageSize int64, initPageSize int64, mapper Mapper) (interface{}, int64, error) {
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
	if mapper == nil {
		return results, count, nil
	}
	r2, er3 := mapper.DbToModels(ctx, results)
	return r2, count, er3
}
