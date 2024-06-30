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
}

func NewBatchUpdaterWithId[T any](database *mongo.Database, collectionName string, options ...func(*T)) *BatchUpdater[T] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	idx := FindIdField(modelType)
	var mp func(*T)
	if len(options) > 0 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &BatchUpdater[T]{collection, idx, mp}
}

func NewBatchUpdater[T any](database *mongo.Database, collectionName string, options ...func(*T)) *BatchUpdater[T] {
	return NewBatchUpdaterWithId[T](database, collectionName, options...)
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
	}
	return failIndices, err
}
