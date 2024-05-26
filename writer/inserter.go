package writer

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Inserter[T any] struct {
	collection *mongo.Collection
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewInserter[T any](database *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *Inserter[T] {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &Inserter[T]{collection: collection, Map: mp}
}

func (w *Inserter[T]) Write(ctx context.Context, model T) error {
	var err error
	if w.Map != nil {
		m2, er0 := w.Map(ctx, model)
		if er0 != nil {
			return er0
		}
		_, err = w.collection.InsertOne(ctx, m2)
		return err
	}
	_, err = w.collection.InsertOne(ctx, model)
	return err
}
