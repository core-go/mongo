package mongo

import (
	"context"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
)

type DefaultSearchResultBuilder struct {
	Database     *mongo.Database
	QueryBuilder QueryBuilder
	SortBuilder  SortBuilder
}

func (b *DefaultSearchResultBuilder) BuildSearchResult(ctx context.Context, collection *mongo.Collection, m interface{}, modelType reflect.Type) (*search.SearchResult, error) {
	query := b.QueryBuilder.BuildQuery(m, modelType)

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
	return b.Build(ctx, collection, modelType, query, sort, searchModel.PageIndex, searchModel.PageSize)
}

func (b *DefaultSearchResultBuilder) Build(ctx context.Context, collection *mongo.Collection, modelType reflect.Type, query bson.M, sort bson.M, pageIndex int, pageSize int) (*search.SearchResult, error) {
	optionsFind := options.Find()
	optionsFind.SetSkip(int64(pageSize * (pageIndex - 1)))
	optionsFind.SetLimit(int64(pageSize))
	if sort != nil {
		optionsFind.SetSort(sort)
	}

	databaseQuery, errFind := collection.Find(ctx, query, optionsFind)
	if errFind != nil {
		return nil, errFind
	}

	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	results := reflect.New(modelsType).Interface()
	errAll := databaseQuery.All(ctx, results)
	if errAll != nil {
		log.Println(errAll)
	}

	var count int
	options := options.Count()
	countDB, errCount := collection.CountDocuments(ctx, query, options)
	if errCount != nil {
		count = 0
	}
	count = int(countDB)

	searchResult := search.SearchResult{}
	searchResult.ItemTotal = count

	searchResult.LastPage = false
	lengthModels := reflect.Indirect(reflect.ValueOf(results)).Len()
	if pageSize*pageIndex+lengthModels >= count {
		searchResult.LastPage = true
	}

	searchResult.Results = results

	return &searchResult, nil
}
