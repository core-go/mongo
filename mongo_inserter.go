package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoInserter struct {
	collection *mongo.Collection
}

func NewMongoInserter(database *mongo.Database, collectionName string) *MongoInserter {
	collection := database.Collection(collectionName)
	return &MongoInserter{collection}
}

func (w *MongoInserter) Write(ctx context.Context, model interface{}) error {
	code, err := InsertOne(ctx, w.collection, model)
	if code == -2 || code == -1 {
		return errors.New("duplicate key")
	}
	return err
}
