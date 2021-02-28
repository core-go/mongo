package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

type Inserter struct {
	collection *mongo.Collection
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewInserter(database *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *Inserter {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	return &Inserter{collection: collection, Map: mp}
}

func (w *Inserter) Write(ctx context.Context, model interface{}) error {
	var code int64
	var err error
	if w.Map != nil {
		m2, er0 := w.Map(ctx, model)
		if er0 != nil {
			return er0
		}
		code, err = InsertOne(ctx, w.collection, m2)
	}
	code, err = InsertOne(ctx, w.collection, model)
	if code == 0 {
		return errors.New("duplicate key")
	}
	return err
}
