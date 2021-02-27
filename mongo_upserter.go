package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type MongoUpserter struct {
	collection *mongo.Collection
	IdName     string
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewMongoWriterById(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string, options ...func(context.Context, interface{}) (interface{}, error)) *MongoUpserter {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	if len(fieldName) == 0 {
		_, idName := FindIdField(modelType)
		fieldName = idName
	}
	return &MongoUpserter{collection: collection, IdName: fieldName, Map: mp}
}

func NewMongoUpserter(database *mongo.Database, collectionName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *MongoUpserter {
	return NewMongoWriterById(database, collectionName, modelType, "", options...)
}

func (w *MongoUpserter) Write(ctx context.Context, model interface{}) error {
	err := Upsert(ctx, w.collection, model, w.IdName)
	return err
}
