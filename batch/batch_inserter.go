package batch

import (
	"context"
	mgo "github.com/core-go/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type BatchInserter[T any] struct {
	collection *mongo.Collection
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewBatchInserter[T any](database *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *BatchInserter[T] {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &BatchInserter[T]{collection: collection, Map: mp}
}

func (w *BatchInserter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	failIndices := make([]int, 0)
	if w.Map != nil {
		_, er0 := mgo.MapModels(ctx, models, w.Map)
		if er0 != nil {
			return failIndices, er0
		}
	}
	return InsertMany[T](ctx, w.collection, models)
}
