package query

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	mgo "github.com/core-go/mongo"
)

type Loader[T any, K any] struct {
	Collection *mongo.Collection
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
	ModelType  reflect.Type
	jsonIdName string
	idIndex    int
	idObjectId bool
}

func UseMongoLoad[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) func(context.Context, K) (*T, error) {
	l := NewMongoLoader[T, K](db, collectionName, idObjectId, options...)
	return l.Load
}
func UseLoad[T any, K any](db *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) func(context.Context, K) (*T, error) {
	l := NewMongoLoader[T, K](db, collectionName, false, options...)
	return l.Load
}
func NewMongoLoader[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) *Loader[T, K] {
	var t T
	modelType := reflect.TypeOf(t)
	idIndex, _, jsonIdName := mgo.FindIdField(modelType)
	if idIndex < 0 {
		log.Println(modelType.Name() + " loader can't use functions that need Id value (Ex Load, Exist, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) > 0 {
		mp = options[0]
	}
	return &Loader[T, K]{db.Collection(collectionName), mp, modelType, jsonIdName, idIndex, idObjectId}
}

func NewLoader[T any, K any](db *mongo.Database, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *Loader[T, K] {
	return NewMongoLoader[T, K](db, collectionName, false, options...)
}
func (a *Loader[T, K]) All(ctx context.Context) ([]T, error) {
	filter := bson.M{}
	cursor, err := a.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var objs []T
	err = cursor.All(ctx, &objs)
	if err != nil {
		return nil, err
	}
	if len(objs) > 0 && a.Map != nil {
		mgo.MapModels(ctx, objs, a.Map)
		return objs, err
	}
	return objs, nil
}

func (a *Loader[T, K]) Load(ctx context.Context, id K) (*T, error) {
	var res T

	if a.idObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return nil, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
		if ok && er0 == nil && a.Map != nil {
			_, er2 := a.Map(ctx, &res)
			if er2 != nil {
				return &res, er2
			}
		}
		return &res, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
	if ok && er2 == nil && a.Map != nil {
		_, er3 := a.Map(ctx, &res)
		if er3 != nil {
			return &res, er3
		}
	}
	return &res, er2
}
func (a *Loader[T, K]) Exist(ctx context.Context, id interface{}) (bool, error) {
	return mgo.Exist(ctx, a.Collection, id, a.idObjectId)
}
