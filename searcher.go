package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Searcher struct {
	modelType     reflect.Type
	collection    *mongo.Collection
	searchBuilder SearchResultBuilder
}

func NewSearcher(db *mongo.Database, modelType reflect.Type, collectionName string, searchBuilder SearchResultBuilder) *Searcher {
	return &Searcher{modelType, db.Collection(collectionName), searchBuilder}
}

func (s *Searcher) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.searchBuilder.BuildSearchResult(ctx, s.collection, m, s.modelType)
}
