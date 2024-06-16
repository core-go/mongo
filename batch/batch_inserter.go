package batch

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
)

type BatchInserter[T any] struct {
	collection *mongo.Collection
	Map        func(*T)
}

func NewBatchInserter[T any](database *mongo.Database, collectionName string, options ...func(*T)) *BatchInserter[T] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	var mp func(*T)
	if len(options) > 0 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &BatchInserter[T]{collection: collection, Map: mp}
}

func (w *BatchInserter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	if w.Map != nil {
		l := len(models)
		for i := 0; i < l; i++ {
			w.Map(&models[i])
		}
	}
	return InsertMany[T](ctx, w.collection, models)
}
