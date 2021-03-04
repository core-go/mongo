package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"reflect"
)

type Loader struct {
	Collection *mongo.Collection
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
	modelType  reflect.Type
	idName     string
	idIndex    int
	idObjectId bool
	keys       []string
}

func NewMongoLoader(db *mongo.Database, collectionName string, modelType reflect.Type, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) *Loader {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	idIndex, idName := FindIdField(modelType)
	if len(idName) == 0 {
		log.Println(modelType.Name() + " loader can't use functions that need Id value (Ex GetById, ExistsById, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	var idNames []string
	idNames = append(idNames, idName)
	return &Loader{db.Collection(collectionName), mp, modelType, idName, idIndex, idObjectId, idNames}
}

func NewLoader(db *mongo.Database, collectionName string, modelType reflect.Type, options ...func(context.Context, interface{}) (interface{}, error)) *Loader {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoLoader(db, collectionName, modelType, false, mp)
}

func (m *Loader) Keys() []string {
	return m.keys
}

func (m *Loader) All(ctx context.Context) (interface{}, error) {
	modelsType := reflect.Zero(reflect.SliceOf(m.modelType)).Type()
	result := reflect.New(modelsType).Interface()
	v, err := FindAndDecode(ctx, m.Collection, bson.M{}, result)
	if v {
		if m.Map != nil {
			return MapModels(ctx, result, m.Map)
		}
		return result, err
	}
	return nil, err
}

func (m *Loader) Load(ctx context.Context, id interface{}) (interface{}, error) {
	r, er1 := FindOneWithId(ctx, m.Collection, id, m.idObjectId, m.modelType)
	if er1 != nil {
		return r, er1
	}
	if m.Map != nil {
		r2, er2 := m.Map(ctx, r)
		if er2 != nil {
			return r, er2
		}
		return r2, er2
	}
	return r, er1
}

func (m *Loader) LoadAndDecode(ctx context.Context, id interface{}, result interface{}) (bool, error) {
	if m.idObjectId {
		objId := id.(string)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return false, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := FindOneAndDecode(ctx, m.Collection, query, result)
		if ok && er0 == nil && m.Map != nil {
			_, er2 := m.Map(ctx, result)
			if er2 != nil {
				return ok, er2
			}
		}
		return ok, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := FindOneAndDecode(ctx, m.Collection, query, result)
	if ok && er2 == nil && m.Map != nil {
		_, er3 := m.Map(ctx, result)
		if er3 != nil {
			return ok, er3
		}
	}
	return ok, er2
}

func (m *Loader) Exist(ctx context.Context, id interface{}) (bool, error) {
	return Exist(ctx, m.Collection, id, m.idObjectId)
}
