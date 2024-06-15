package adapter

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mgo "github.com/core-go/mongo"
)

type Mapper[T any] interface {
	DbToModel(*T)
	ModelToDb(*T)
	MapToDb(map[string]interface{}) map[string]interface{}
}
type Adapter[T any, K any] struct {
	Collection   *mongo.Collection
	ModelType    reflect.Type
	jsonIdName   string
	idIndex      int
	ObjectId     bool
	Map          map[string]string
	versionField string
	versionIndex int
	Mapper       Mapper[T]
}

func NewMongoAdapterWithVersion[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, versionField string, options ...Mapper[T]) *Adapter[T, K] {
	var mapper Mapper[T]
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
		log.Println(modelType.Name() + " Adapter can't use functions that need Id value (Ex Load, Exist, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	if len(versionField) > 0 {
		index := mgo.FindFieldIndex(modelType, versionField)
		if index >= 0 {
			return &Adapter[T, K]{Collection: db.Collection(collectionName), ModelType: modelType, jsonIdName: jsonIdName, idIndex: idIndex, ObjectId: idObjectId,
				Map: mgo.MakeBsonMap(modelType), versionField: versionField, versionIndex: index, Mapper: mapper}
		}
	}
	return &Adapter[T, K]{Collection: db.Collection(collectionName), ModelType: modelType, jsonIdName: jsonIdName, idIndex: idIndex, ObjectId: idObjectId,
		Map: mgo.MakeBsonMap(modelType), Mapper: mapper, versionIndex: -1}
}
func NewAdapterWithVersion[T any, K any](db *mongo.Database, collectionName string, versionField string, options ...Mapper[T]) *Adapter[T, K] {
	return NewMongoAdapterWithVersion[T, K](db, collectionName, false, versionField, options...)
}
func NewAdapter[T any, K any](db *mongo.Database, collectionName string, options ...Mapper[T]) *Adapter[T, K] {
	return NewMongoAdapterWithVersion[T, K](db, collectionName, false, "", options...)
}
func (a *Adapter[T, K]) All(ctx context.Context) ([]T, error) {
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
	if a.Mapper != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			a.Mapper.DbToModel(&objs[i])
		}
	}
	return objs, nil
}
func (a *Adapter[T, K]) Load(ctx context.Context, id K) (*T, error) {
	var res T
	if a.ObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return nil, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
		if ok && er0 == nil && a.Mapper != nil {
			a.Mapper.DbToModel(&res)
		}
		return &res, er0
	}
	query := bson.M{"_id": id}
	ok, er2 := mgo.FindOneAndDecode(ctx, a.Collection, query, &res)
	if er2 != nil {
		return nil, er2
	}
	if !ok {
		return nil, nil
	}
	if a.Mapper != nil {
		a.Mapper.DbToModel(&res)
	}
	return &res, er2
}

func (a *Adapter[T, K]) Exist(ctx context.Context, id K) (bool, error) {
	return mgo.Exist(ctx, a.Collection, id, a.ObjectId)
}
func (a *Adapter[T, K]) Create(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	if a.versionIndex >= 0 {
		setVersion(model, a.versionIndex)
	}
	res, id, err := InsertOne(ctx, a.Collection, model)
	if err != nil {
		return res, err
	}
	if id != nil {
		vo := reflect.Indirect(reflect.ValueOf(model))
		idF := vo.Field(a.idIndex)
		switch idF.Kind() {
		case reflect.String:
			idF.Set(reflect.ValueOf(id.Hex()))
		default:
			idF.Set(reflect.ValueOf(id))
		}
	}
	return res, err
}
func (a *Adapter[T, K]) Update(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	if a.versionIndex >= 0 {
		return mgo.UpdateByIdAndVersion(ctx, a.Collection, model, a.versionIndex)
	}
	idQuery := mgo.BuildQueryByIdFromObject(model)
	return mgo.UpdateOne(ctx, a.Collection, model, idQuery)
}
func (a *Adapter[T, K]) Patch(ctx context.Context, model map[string]interface{}) (int64, error) {
	if a.Mapper != nil {
		m3 := a.Mapper.MapToDb(model)
		if a.versionIndex >= 0 {
			return mgo.PatchByIdAndVersion(ctx, a.Collection, m3, a.Map, a.jsonIdName, a.versionField)
		}
		idQuery := mgo.BuildQueryByIdFromMap(m3, a.jsonIdName)
		b0 := mgo.MapToBson(m3, a.Map)
		return mgo.PatchOne(ctx, a.Collection, b0, idQuery)
	}
	if a.versionIndex >= 0 {
		return mgo.PatchByIdAndVersion(ctx, a.Collection, model, a.Map, a.jsonIdName, a.versionField)
	}
	idQuery := mgo.BuildQueryByIdFromMap(model, a.jsonIdName)
	b := mgo.MapToBson(model, a.Map)
	return mgo.PatchOne(ctx, a.Collection, b, idQuery)
}
func (a *Adapter[T, K]) Save(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	if a.versionIndex >= 0 {
		return mgo.UpsertOneWithVersion(ctx, a.Collection, model, a.versionIndex)
	}
	idQuery := mgo.BuildQueryByIdFromObject(model)
	return mgo.UpsertOne(ctx, a.Collection, idQuery, model)
}
func (a *Adapter[T, K]) Delete(ctx context.Context, id K) (int64, error) {
	if a.ObjectId {
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

func setVersion(model interface{}, versionIndex int) bool {
	modelType := reflect.TypeOf(model).Elem()
	versionType := modelType.Field(versionIndex).Type.String()
	vo := reflect.Indirect(reflect.ValueOf(model))
	switch versionType {
	case "int":
		vo.Field(versionIndex).Set(reflect.ValueOf(int(1)))
		return true
	case "int32":
		vo.Field(versionIndex).Set(reflect.ValueOf(int32(1)))
		return true
	case "int64":
		vo.Field(versionIndex).Set(reflect.ValueOf(int64(1)))
		return true
	default:
		return false
	}
}
func InsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (int64, *primitive.ObjectID, error) {
	result, err := collection.InsertOne(ctx, model)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key error collection:") {
			return 0, nil, nil
		} else {
			return 0, nil, err
		}
	} else {
		if idValue, ok := result.InsertedID.(primitive.ObjectID); ok {
			return 1, &idValue, nil
		}
		return 1, nil, nil
	}
}
