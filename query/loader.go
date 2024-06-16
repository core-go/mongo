package query

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	mgo "github.com/core-go/mongo"
)

type Loader[T any, K any] struct {
	Collection *mongo.Collection
	ObjectId   bool
	idIndex    int
	idJson     string
	Map        func(*T)
}

func NewMongoLoader[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, options ...func(*T)) *Loader[T, K] {
	var mapper func(*T)
	if len(options) > 0 {
		mapper = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	idIndex, _, jsonIdName := mgo.FindIdField(modelType)
	if idIndex < 0 {
		panic(modelType.Name() + " Loader can't use functions that need Id value (Ex Load, Exist) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	adapter := &Loader[T, K]{Collection: db.Collection(collectionName), idJson: jsonIdName, idIndex: idIndex, ObjectId: idObjectId, Map: mapper}
	return adapter
}
func NewLoader[T any, K any](db *mongo.Database, collectionName string, options ...func(*T)) *Loader[T, K] {
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
	if a.Map != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			a.Map(&objs[i])
		}
	}
	return objs, nil
}
func (a *Loader[T, K]) Load(ctx context.Context, id K) (*T, error) {
	var res T
	if a.ObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return nil, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := mgo.FindOne(ctx, a.Collection, query, &res)
		if ok && er0 == nil && a.Map != nil {
			a.Map(&res)
		}
		return &res, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := mgo.FindOne(ctx, a.Collection, query, &res)
	if er2 != nil {
		return nil, er2
	}
	if !ok {
		return nil, nil
	}
	if a.Map != nil {
		a.Map(&res)
	}
	return &res, er2
}

func (a *Loader[T, K]) Exist(ctx context.Context, id K) (bool, error) {
	if a.ObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return false, err
		}
		return mgo.Exist(ctx, a.Collection, objectId)
	}
	return mgo.Exist(ctx, a.Collection, id)
}
