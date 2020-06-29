package mongo

import (
	"context"
	"github.com/common-go/search"
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

func (s *SearchService) Search(ctx context.Context, m interface{}) (*search.SearchResult, error) {
	return s.searchBuilder.BuildSearchResult(ctx, s.collection, m, s.modelType)
}
