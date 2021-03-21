package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Upserter struct {
	collection *mongo.Collection
	IdName     string
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
}

func NewUpserterById(database *mongo.Database, collectionName string, modelType reflect.Type, fieldName string, options ...func(context.Context, interface{}) (interface{}, error)) *Upserter {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	collection := database.Collection(collectionName)
	if len(fieldName) == 0 {
		_, idName, _ := FindIdField(modelType)
		fieldName = idName
	}
	return &Upserter{collection: collection, IdName: fieldName, Map: mp}
}

func NewUpserter(database *mongo.Database, collectionName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *Upserter {
	return NewUpserterById(database, collectionName, modelType, "", options...)
}

func (w *Upserter) Write(ctx context.Context, model interface{}) error {
	if w.Map != nil {
		m2, er0 := w.Map(ctx, model)
		if er0 != nil {
			return er0
		}
		return Upsert(ctx, w.collection, m2, w.IdName)
	}
	err := Upsert(ctx, w.collection, model, w.IdName)
	return err
}
