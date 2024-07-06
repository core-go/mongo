package batch

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
)

type BatchInserter[T any] struct {
	collection *mongo.Collection
	Map        func(*T)
	retryAll   bool
}

func NewBatchInserterWithRetry[T any](db *mongo.Database, collectionName string, retryAll bool, opts ...func(*T)) *BatchInserter[T] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	var mp func(*T)
	if len(opts) > 0 {
		mp = opts[0]
	}
	collection := db.Collection(collectionName)
	return &BatchInserter[T]{collection: collection, Map: mp, retryAll: retryAll}
}
func NewBatchInserter[T any](db *mongo.Database, collectionName string, opts ...func(*T)) *BatchInserter[T] {
	return NewBatchInserterWithRetry[T](db, collectionName, false, opts...)
}

func (w *BatchInserter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	if w.Map != nil {
		l := len(models)
		for i := 0; i < l; i++ {
			w.Map(&models[i])
		}
	}
	fails, err := InsertMany[T](ctx, w.collection, models)
	if err != nil && len(fails) == 0 && w.retryAll {
		l := len(models)
		failIndices := make([]int, 0)
		for i := 0; i < l; i++ {
			failIndices = append(failIndices, i)
		}
		return failIndices, err
	}
	return fails, err
}
