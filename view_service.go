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
	Map        func(ctx context.Context, model interface{}) (interface{}, error)
	modelType  reflect.Type
	idName     string
	idIndex    int
	idObjectId bool
	keys       []string
}

func NewMongoViewService(db *mongo.Database, modelType reflect.Type, collectionName string, idObjectId bool, options ...func(context.Context, interface{}) (interface{}, error)) *ViewService {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	idIndex, idName := FindIdField(modelType)
	if len(idName) == 0 {
		log.Println(modelType.Name() + " repository can't use functions that need Id value (Ex GetById, ExistsById, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	var idNames []string
	idNames = append(idNames, idName)
	return &ViewService{db.Collection(collectionName), mp, modelType, idName, idIndex, idObjectId, idNames}
}

func NewViewService(db *mongo.Database, modelType reflect.Type, collectionName string, options ...func(context.Context, interface{}) (interface{}, error)) *ViewService {
	var mp func(context.Context, interface{}) (interface{}, error)
	if len(options) >= 1 {
		mp = options[0]
	}
	return NewMongoViewService(db, modelType, collectionName, false, mp)
}

func (m *ViewService) Keys() []string {
	return m.keys
}

func (m *ViewService) All(ctx context.Context) (interface{}, error) {
	modelsType := reflect.Zero(reflect.SliceOf(m.modelType)).Type()
	result := reflect.New(modelsType).Interface()
	v, err := FindAndDecode(ctx, m.Collection, bson.M{}, result)
	if v {
		if m.Map != nil {
			valueModelObject := reflect.Indirect(reflect.ValueOf(result))
			if valueModelObject.Kind() == reflect.Ptr {
				valueModelObject = reflect.Indirect(valueModelObject)
			}
			if valueModelObject.Kind() == reflect.Slice {
				for i := 0; i < valueModelObject.Len(); i++ {
					m.Map(ctx, valueModelObject.Index(i))
				}
			}
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
	if m.Map != nil {
		r2, er2 := m.Map(ctx, r)
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

func (m *ViewService) Exist(ctx context.Context, id interface{}) (bool, error) {
	return Exist(ctx, m.Collection, id, m.idObjectId)
}
