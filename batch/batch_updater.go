package batch

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type BatchUpdater[T any] struct {
	collection *mongo.Collection
	Idx        int
	Map        func(*T)
	retryAll   bool
}

func NewBatchUpdaterWithRetry[T any](db *mongo.Database, collectionName string, retryAll bool, opts ...func(*T)) *BatchUpdater[T] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	idx := FindIdField(modelType)
	if idx < 0 {
		panic("T must contain Id field, which has '_id' bson tag")
	}
	var mp func(*T)
	if len(opts) > 0 {
		mp = opts[0]
	}
	collection := db.Collection(collectionName)
	return &BatchUpdater[T]{collection, idx, mp, retryAll}
}
func NewBatchUpdater[T any](db *mongo.Database, collectionName string, retryAll bool, opts ...func(*T)) *BatchUpdater[T] {
	return NewBatchUpdaterWithRetry[T](db, collectionName, false, opts...)
}
func (w *BatchUpdater[T]) Write(ctx context.Context, models []T) ([]int, error) {
	failIndices := make([]int, 0)
	var err error
	if w.Map != nil {
		l := len(models)
		for i := 0; i < l; i++ {
			w.Map(&models[i])
		}
	}
	_, err = UpdateMany[T](ctx, w.collection, models, w.Idx)
	if err == nil {
		return failIndices, err
	}

	if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
		for _, writeError := range bulkWriteException.WriteErrors {
			failIndices = append(failIndices, writeError.Index)
		}
	} else if w.retryAll {
		l := len(models)
		fails := make([]int, 0)
		for i := 0; i < l; i++ {
			fails = append(fails, i)
		}
		return fails, err
	}
	return failIndices, err
}
