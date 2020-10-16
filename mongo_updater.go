package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type MongoUpdater struct {
	collection *mongo.Collection
	IdName     string
	modelType  reflect.Type
}

func NewMongoUpdaterWithIdName(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string) *MongoUpdater {
	if len(fieldName) == 0 {
		_, idName := FindIdField(modelType)
		fieldName = idName
	}
	collection := database.Collection(collectionName)
	return &MongoUpdater{collection, fieldName, modelType}
}

func NewMongoUpdater(database *mongo.Database, collectionName string, modelType reflect.Type) *MongoUpdater {
	return NewMongoUpdaterWithIdName(database, collectionName, modelType, "")
}

func (w *MongoUpdater) Write(ctx context.Context, model interface{}) error {
	err := Update(ctx, w.collection, model, w.IdName)
	return err
}
