package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type SearchService struct {
	modelType     reflect.Type
	collection    *mongo.Collection
	searchBuilder SearchResultBuilder
}

func NewSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder) *SearchService {
	return &SearchService{modelType, db.Collection(collectionName), searchBuilder}
}

func (s *SearchService) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.searchBuilder.BuildSearchResult(ctx, s.collection, m, s.modelType)
}
