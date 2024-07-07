package batch

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type StreamUpdater[T any] struct {
	collection *mongo.Collection
	Idx        int
	batchSize  int
	batch      []interface{}
	Map        func(T)
	isPointer  bool
}

func NewStreamUpdater[T any](db *mongo.Database, collectionName string, batchSize int, opts ...func(T)) *StreamUpdater[T] {
	var t T
	modelType := reflect.TypeOf(t)
	isPointer := false
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
		isPointer = true
	}
	idx := FindIdField(modelType)
	if idx < 0 {
		panic("T must contain Id field, which has '_id' bson tag")
	}
	var mp func(T)
	if len(opts) > 0 {
		mp = opts[0]
	}
	collection := db.Collection(collectionName)
	batch := make([]interface{}, 0)
	return &StreamUpdater[T]{collection, idx, batchSize, batch, mp, isPointer}
}
func (w *StreamUpdater[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		w.Map(model)
	}
	vo := reflect.ValueOf(model)
	if w.isPointer {
		vo = reflect.Indirect(vo)
	}
	w.batch = append(w.batch, vo.Interface())
	if len(w.batch) >= w.batchSize {
		return w.Flush(ctx)
	}
	return nil
}
func (w *StreamUpdater[T]) Flush(ctx context.Context) error {
	if len(w.batch) == 0 {
		return nil
	}
	_, err := UpdateMany[interface{}](ctx, w.collection, w.batch, w.Idx)
	w.batch = make([]interface{}, 0)
	return err
}
