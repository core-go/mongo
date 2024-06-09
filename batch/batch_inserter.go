package batch

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type BatchInserter[T any] struct {
	collection *mongo.Collection
	Map        func(T) T
}

func NewBatchInserter[T any](database *mongo.Database, collectionName string, options ...func(T) T) *BatchInserter[T] {
	var mp func(T) T
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &BatchInserter[T]{collection: collection, Map: mp}
}

func (w *BatchInserter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	if w.Map != nil {
		list := make([]T, 0)
		l := len(models)
		for i := 0; i < l; i++ {
			obj := w.Map(models[i])
			list = append(list, obj)
		}
		return InsertMany[T](ctx, w.collection, list)
	}
	return InsertMany[T](ctx, w.collection, models)
}
