package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

type Writer[T any] struct {
	collection *mongo.Collection
	idIndex    int
	Map        func(T)
	isPointer  bool
}

func NewWriter[T any](database *mongo.Database, collectionName string, options ...func(T)) *Writer[T] {
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
	return &Writer[T]{collection: collection, idIndex: index, Map: mp, isPointer: isPointer}
}

func (w *Writer[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		w.Map(model)
	}
	vo := reflect.ValueOf(model)
	if w.isPointer {
		vo = reflect.Indirect(vo)
	}
	id := vo.Field(w.idIndex).Interface()
	sid, ok := id.(string)
	if ok && len(sid) == 0 || isNil(id) {
		_, err := w.collection.InsertOne(ctx, model)
		return err
	}
	return Upsert(ctx, w.collection, id, model)
}

func Upsert(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) error {
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, updateQuery, opts)
	return err
}
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
