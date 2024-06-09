package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
)

type Writer[T any] struct {
	collection *mongo.Collection
	IdName     string
	Map        func(T) T
}

func NewWriterById[T any](database *mongo.Database, collectionName string, fieldName string, options ...func(T) T) *Writer[T] {
	var mp func(T) T
	if len(options) >= 1 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	collection := database.Collection(collectionName)
	if len(fieldName) == 0 {
		_, idName, _ := mgo.FindIdField(modelType)
		fieldName = idName
	}
	return &Writer[T]{collection: collection, IdName: fieldName, Map: mp}
}

func NewWriter[T any](database *mongo.Database, collectionName string, options ...func(T) T) *Writer[T] {
	return NewWriterById[T](database, collectionName, "", options...)
}

func (w *Writer[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		m2 := w.Map(model)
		return mgo.Upsert(ctx, w.collection, m2, w.IdName)
	}
	err := mgo.Upsert(ctx, w.collection, model, w.IdName)
	return err
}