package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type MongoWriter struct {
	collection *mongo.Collection
	IdName     string
}

func NewMongoWriterById(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string) *MongoWriter {
	collection := database.Collection(collectionName)
	if len(fieldName) == 0 {
		_, idName := FindIdField(modelType)
		fieldName = idName
	}
	return &MongoWriter{collection, fieldName}
}

func NewMongoWriter(database *mongo.Database, collectionName string, modelType reflect.Type) *MongoWriter {
	return NewMongoWriterById(database, collectionName, modelType, "")
}

func (w *MongoWriter) Write(ctx context.Context, model interface{}) error {
	err := Upsert(ctx, w.collection, model, w.IdName)
	return err
}
