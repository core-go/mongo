package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"reflect"
)

type ViewService struct {
	Collection *mongo.Collection
	Mapper     Mapper
	modelType  reflect.Type
	idName     string
	idIndex    int
	idObjectId bool
	keys       []string
}

func NewViewService(db *mongo.Database, modelType reflect.Type, collectionName string, idObjectId bool, mapper Mapper) *ViewService {
	idIndex, idName := FindIdField(modelType)
	if len(idName) == 0 {
		log.Println(modelType.Name() + " repository can't use functions that need Id value (Ex GetById, ExistsById, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	var idNames []string
	idNames = append(idNames, idName)
	return &ViewService{db.Collection(collectionName), mapper, modelType, idName, idIndex, idObjectId, idNames}
}

func NewDefaultViewService(db *mongo.Database, modelType reflect.Type, collectionName string) *ViewService {
	return NewViewService(db, modelType, collectionName, false, nil)
}

func (m *ViewService) Keys() []string {
	return m.keys
}

func (m *ViewService) All(ctx context.Context) (interface{}, error) {
	modelsType := reflect.Zero(reflect.SliceOf(m.modelType)).Type()
	result := reflect.New(modelsType).Interface()
	v, err := FindAndDecode(ctx, m.Collection, bson.M{}, result)
	if v {
		if m.Mapper != nil {
			r2, er2 := m.Mapper.DbToModels(ctx, result)
			if er2 != nil {
				return result, err
			}
			return r2, err
		}
		return result, err
	}
	return nil, err
}

func (m *ViewService) Load(ctx context.Context, id interface{}) (interface{}, error) {
	r, er1 := FindOneWithId(ctx, m.Collection, id, m.idObjectId, m.modelType)
	if er1 != nil {
		return r, er1
	}
	if m.Mapper != nil {
		r2, er2 := m.Mapper.DbToModel(ctx, r)
		if er2 != nil {
			return r, er2
		}
		return r2, er2
	}
	return r, er1
}

func (m *ViewService) LoadAndDecode(ctx context.Context, id interface{}, result interface{}) (bool, error) {
	if m.idObjectId {
		objId := id.(string)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return false, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := FindOneAndDecode(ctx, m.Collection, query, result)
		if ok && er0 == nil && m.Mapper != nil {
			_, er2 := m.Mapper.DbToModel(ctx, result)
			if er2 != nil {
				return ok, er2
			}
		}
		return ok, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := FindOneAndDecode(ctx, m.Collection, query, result)
	if ok && er2 == nil && m.Mapper != nil {
		_, er3 := m.Mapper.DbToModel(ctx, result)
		if er3 != nil {
			return ok, er3
		}
	}
	return ok, er2
}

func (m *ViewService) Exist(ctx context.Context, id interface{}) (bool, error) {
	return Exist(ctx, m.Collection, id, m.idObjectId)
}
