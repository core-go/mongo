package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Searcher struct {
	modelType  reflect.Type
	collection *mongo.Collection
	Build      func(ctx context.Context, collection *mongo.Collection, searchModel interface{}, modelType reflect.Type) (interface{}, int64, error)
}

func NewSearcher(db *mongo.Database, modelType reflect.Type, collectionName string, build func(context.Context, *mongo.Collection, interface{}, reflect.Type) (interface{}, int64, error)) *Searcher {
	return &Searcher{modelType: modelType, collection: db.Collection(collectionName), Build: build}
}

func (s *Searcher) Search(ctx context.Context, m interface{}) (interface{}, int64, error) {
	return s.Build(ctx, s.collection, m, s.modelType)
}
