package batch

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type BatchWriter[T any] struct {
	collection *mongo.Collection
	Idx        int
	Map        func(*T)
}

func NewBatchWriter[T any](database *mongo.Database, collectionName string, options ...func(*T)) *BatchWriter[T] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	var mp func(*T)
	if len(options) > 0 {
		mp = options[0]
	}
	idx := FindIdField(modelType)
	collection := database.Collection(collectionName)
	return &BatchWriter[T]{collection, idx, mp}
}
func (w *BatchWriter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	failIndices := make([]int, 0)
	var err error
	if w.Map != nil {
		l := len(models)
		for i := 0; i < l; i++ {
			w.Map(&models[i])
		}
	}
	_, err = UpsertMany[T](ctx, w.collection, models, w.Idx)

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
