package mongo

import (
	"context"
	"github.com/common-go/search"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type DefaultSearchService struct {
	Database      *mongo.Database
	modelType     reflect.Type
	collection    *mongo.Collection
	searchBuilder SearchResultBuilder
}

func NewDefaultSearchService(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder) *DefaultSearchService {
	return &DefaultSearchService{db, modelType, db.Collection(collectionName), searchBuilder}
}

func (s *DefaultSearchService) Search(ctx context.Context, m interface{}) (*search.SearchResult, error) {
	return s.searchBuilder.BuildSearchResult(ctx, s.collection, m, s.modelType)
}
