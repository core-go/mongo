package batch

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
)

type BatchWriter[T any] struct {
	collection *mongo.Collection
	Idx        int
	Map        func(T) T
}

func NewBatchWriterWithId[T any](database *mongo.Database, collectionName string, options ...func(T) T) *BatchWriter[T] {
	var mp func(T) T
	if len(options) >= 1 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	idx, _, _ := mgo.FindIdField(modelType)
	collection := database.Collection(collectionName)
	return &BatchWriter[T]{collection, idx, mp}
}
func NewBatchWriter[T any](database *mongo.Database, collectionName string, options ...func(T) T) *BatchWriter[T] {
	return NewBatchWriterWithId[T](database, collectionName, options...)
}
func (w *BatchWriter[T]) Write(ctx context.Context, models []T) ([]int, error) {
	failIndices := make([]int, 0)
	var err error
	if w.Map != nil {
		l := len(models)
		for i := 0; i < l; i++ {
			models[i] = w.Map(models[i])
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
	} else {
		l := len(models)
		for i := 0; i < l; i++ {
			failIndices = append(failIndices, i)
		}
	}
	return failIndices, err
}
