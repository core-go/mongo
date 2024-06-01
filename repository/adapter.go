package repository

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mgo "github.com/core-go/mongo"
)

type Repository[T any, K any] struct {
	Collection   *mongo.Collection
	ModelType    reflect.Type
	jsonIdName   string
	idIndex      int
	idObjectId   bool
	Map          map[string]string
	versionField string
	versionIndex int
	Mapper       mgo.Mapper
}

func NewMongoRepositoryWithVersion[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, versionField string, options ...mgo.Mapper) *Repository[T, K] {
	var mapper mgo.Mapper
	if len(options) > 0 {
		mapper = options[0]
	}
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	idIndex, _, jsonIdName := mgo.FindIdField(modelType)
	if idIndex < 0 {
		log.Println(modelType.Name() + " loader can't use functions that need Id value (Ex Load, Exist, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	if len(versionField) > 0 {
		index := mgo.FindFieldIndex(modelType, versionField)
		if index >= 0 {
			return &Repository[T, K]{Collection: db.Collection(collectionName), ModelType: modelType, jsonIdName: jsonIdName, idIndex: idIndex, idObjectId: idObjectId,
				Map: mgo.MakeBsonMap(modelType), versionField: versionField, versionIndex: index, Mapper: mapper}
		}
	}
	return &Repository[T, K]{Collection: db.Collection(collectionName), ModelType: modelType, jsonIdName: jsonIdName, idIndex: idIndex, idObjectId: idObjectId,
		Map: mgo.MakeBsonMap(modelType), Mapper: mapper}
}
func NewRepositoryWithVersion[T any, K any](db *mongo.Database, collectionName string, versionField string, options ...mgo.Mapper) *Repository[T, K] {
	return NewMongoRepositoryWithVersion[T, K](db, collectionName, false, versionField, options...)
}
func NewRepository[T any, K any](db *mongo.Database, collectionName string, options ...mgo.Mapper) *Repository[T, K] {
	return NewMongoRepositoryWithVersion[T, K](db, collectionName, false, "", options...)
}
func (a *Repository[T, K]) All(ctx context.Context) ([]T, error) {
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
	if len(objs) > 0 && a.Mapper != nil {
		mgo.MapModels(ctx, objs, a.Mapper.DbToModel)
		return objs, err
	}
	return objs, nil
}
func (a *Repository[T, K]) Load(ctx context.Context, id K) (*T, error) {
	var res T

	if a.idObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return nil, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
		if ok && er0 == nil && a.Mapper != nil {
			_, er2 := a.Mapper.DbToModel(ctx, &res)
			if er2 != nil {
				return &res, er2
			}
		}
		return &res, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
	if ok && er2 == nil && a.Mapper != nil {
		_, er3 := a.Mapper.DbToModel(ctx, &res)
		if er3 != nil {
			return &res, er3
		}
	}
	return &res, er2
}

func (a *Repository[T, K]) Exist(ctx context.Context, id K) (bool, error) {
	return mgo.Exist(ctx, a.Collection, id, a.idObjectId)
}
func (a *Repository[T, K]) Create(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		m2, err := a.Mapper.ModelToDb(ctx, model)
		if err != nil {
			return 0, err
		}
		if a.versionIndex >= 0 {
			return mgo.InsertOneWithVersion(ctx, a.Collection, m2, a.versionIndex)
		}
		return mgo.InsertOne(ctx, a.Collection, m2)
	}
	if a.versionIndex >= 0 {
		return mgo.InsertOneWithVersion(ctx, a.Collection, model, a.versionIndex)
	}
	return mgo.InsertOne(ctx, a.Collection, model)
}
func (a *Repository[T, K]) Update(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		m2, err := a.Mapper.ModelToDb(ctx, model)
		if err != nil {
			return 0, err
		}
		if a.versionIndex >= 0 {
			return mgo.UpdateByIdAndVersion(ctx, a.Collection, m2, a.versionIndex)
		}
		idQuery := mgo.BuildQueryByIdFromObject(m2)
		return mgo.UpdateOne(ctx, a.Collection, m2, idQuery)
	}
	if a.versionIndex >= 0 {
		return mgo.UpdateByIdAndVersion(ctx, a.Collection, model, a.versionIndex)
	}
	idQuery := mgo.BuildQueryByIdFromObject(model)
	return mgo.UpdateOne(ctx, a.Collection, model, idQuery)
}
func (a *Repository[T, K]) Patch(ctx context.Context, model map[string]interface{}) (int64, error) {
	if a.Mapper != nil {
		m2, err := a.Mapper.ModelToDb(ctx, model)
		if err != nil {
			return 0, err
		}
		m3, ok1 := m2.(map[string]interface{})
		if !ok1 {
			return 0, fmt.Errorf("result of LocationToBson must be a map[string]interface{}")
		}
		if a.versionIndex >= 0 {
			return mgo.PatchByIdAndVersion(ctx, a.Collection, m3, a.Map, a.jsonIdName, a.versionField)
		}
		jsonName0 := mgo.GetJsonByIndex(a.ModelType, a.idIndex)
		idQuery := mgo.BuildQueryByIdFromMap(m3, jsonName0)
		b0 := mgo.MapToBson(m3, a.Map)
		return mgo.PatchOne(ctx, a.Collection, b0, idQuery)
	}
	if a.versionIndex >= 0 {
		return mgo.PatchByIdAndVersion(ctx, a.Collection, model, a.Map, a.jsonIdName, a.versionField)
	}
	jsonName := mgo.GetJsonByIndex(a.ModelType, a.idIndex)
	idQuery := mgo.BuildQueryByIdFromMap(model, jsonName)
	b := mgo.MapToBson(model, a.Map)
	return mgo.PatchOne(ctx, a.Collection, b, idQuery)
}
func (a *Repository[T, K]) Save(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		m2, err := a.Mapper.ModelToDb(ctx, model)
		if err != nil {
			return 0, err
		}
		if a.versionIndex >= 0 {
			return mgo.UpsertOneWithVersion(ctx, a.Collection, m2, a.versionIndex)
		}
		idQuery := mgo.BuildQueryByIdFromObject(m2)
		return mgo.UpsertOne(ctx, a.Collection, idQuery, m2)
	}
	if a.versionIndex >= 0 {
		return mgo.UpsertOneWithVersion(ctx, a.Collection, model, a.versionIndex)
	}
	idQuery := mgo.BuildQueryByIdFromObject(model)
	return mgo.UpsertOne(ctx, a.Collection, idQuery, model)
}
func (a *Repository[T, K]) Delete(ctx context.Context, id K) (int64, error) {
	if a.idObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return 0, err
		}
		query := bson.M{"_id": objectId}
		return mgo.DeleteOne(ctx, a.Collection, query)
	}
	query := bson.M{"_id": id}
	return mgo.DeleteOne(ctx, a.Collection, query)
}
