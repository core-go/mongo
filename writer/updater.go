package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Updater[T any] struct {
	collection *mongo.Collection
	idIndex    int
	Map        func(T)
	isPointer  bool
}

func NewUpdaterWithId[T any](database *mongo.Database, collectionName string, options ...func(T)) *Updater[T] {
	var mp func(T)
	if len(options) > 0 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	isPointer := false
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
		isPointer = true
	}
	index := FindIdField(modelType)
	collection := database.Collection(collectionName)
	return &Updater[T]{collection: collection, idIndex: index, Map: mp, isPointer: isPointer}
}

func NewUpdater[T any](database *mongo.Database, collectionName string, options ...func(T)) *Updater[T] {
	return NewUpdaterWithId[T](database, collectionName, options...)
}

func (w *Updater[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		w.Map(model)
	}
	vo := reflect.ValueOf(model)
	if w.isPointer {
		vo = reflect.Indirect(vo)
	}
	id := vo.Field(w.idIndex).Interface()
	return Update(ctx, w.collection, id, model)
}

func Update(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) error { //Patch
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	_, err := collection.UpdateOne(ctx, filter, updateQuery)
	return err
}
