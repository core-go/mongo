package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
)

type MongoWriter[T any] struct {
	collection *mongo.Collection
	IdName     string
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewMongoWriterById[T any](database *mongo.Database, collectionName string, fieldName string, options ...func(context.Context, interface{}) (interface{}, error)) *MongoWriter[T] {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	collection := database.Collection(collectionName)
	if len(fieldName) == 0 {
		_, idName, _ := mgo.FindIdField(modelType)
		fieldName = idName
	}
	return &MongoWriter[T]{collection: collection, IdName: fieldName, Map: mp}
}

func NewMongoWriter[T any](database *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *MongoWriter[T] {
	return NewMongoWriterById[T](database, collectionName, "", options...)
}

func (w *MongoWriter[T]) Write(ctx context.Context, model T) error {
	if w.Map != nil {
		m2, er0 := w.Map(ctx, model)
		if er0 != nil {
			return er0
		}
		return mgo.Upsert(ctx, w.collection, m2, w.IdName)
	}
	err := mgo.Upsert(ctx, w.collection, model, w.IdName)
	return err
}
