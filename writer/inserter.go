package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Inserter[T any] struct {
	collection *mongo.Collection
	Map        func(T)
}

func NewInserter[T any](database *mongo.Database, collectionName string, options ...func(T)) *Inserter[T] {
	var mp func(T)
	if len(options) > 0 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &Inserter[T]{collection: collection, Map: mp}
}

func (w *Inserter[T]) Write(ctx context.Context, model T) error {
	var err error
	if w.Map != nil {
		w.Map(model)
	}
	_, err = w.collection.InsertOne(ctx, model)
	return err
}
