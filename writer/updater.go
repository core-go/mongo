package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
)

type Updater[T any] struct {
	collection *mongo.Collection
	IdName     string
	Map        func(T)
}

func NewUpdaterWithId[T any](database *mongo.Database, collectionName string, fieldName string, options ...func(T)) *Updater[T] {
	var mp func(T)
	if len(options) > 0 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if len(fieldName) == 0 {
		_, idName, _ := mgo.FindIdField(modelType)
		fieldName = idName
	}
	collection := database.Collection(collectionName)
	return &Updater[T]{collection: collection, IdName: fieldName, Map: mp}
}

func NewUpdater[T any](database *mongo.Database, collectionName string, options ...func(T)) *Updater[T] {
	return NewUpdaterWithId[T](database, collectionName, "", options...)
}

func (w *Updater[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		w.Map(model)
	}
	return mgo.Update(ctx, w.collection, model, w.IdName)
}
