package mongo

import (
	"context"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	return BuildSearchResult(ctx, collection, modelType, query, fields, sort, searchModel.PageIndex, searchModel.PageSize, searchModel.FirstPageSize, b.Mapper)
}
