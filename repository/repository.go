package repository

import (
	"context"
	"fmt"
	"github.com/core-go/core/errors"
	"go.mongodb.org/mongo-driver/mongo/options"
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
type Repository[T any, K any] struct {
	Collection   *mongo.Collection
	Map          map[string]string
	ObjectId     bool
	idIndex      int
	idJson       string
	versionIndex int
	versionJson  string
	versionBson  string
	Mapper       Mapper[T]
}

func FindFieldByName(modelType reflect.Type, fieldName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		if field.Name == fieldName {
			name1 := fieldName
			name2 := fieldName
			tag1, ok1 := field.Tag.Lookup("json")
			tag2, ok2 := field.Tag.Lookup("bson")
			if ok1 {
				name1 = strings.Split(tag1, ",")[0]
			}
			if ok2 {
				name2 = strings.Split(tag2, ",")[0]
			}
			return i, name1, name2
		}
	}
	return -1, fieldName, fieldName
}
func NewMongoRepositoryWithVersion[T any, K any](db *mongo.Database, collectionName string, idObjectId bool, versionField string, options ...Mapper[T]) *Repository[T, K] {
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
		log.Println(modelType.Name() + " Repository can't use functions that need Id value (Ex Load, Exist, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	adapter := &Repository[T, K]{Collection: db.Collection(collectionName), idJson: jsonIdName, idIndex: idIndex, ObjectId: idObjectId,
		Map: mgo.MakeBsonMap(modelType), Mapper: mapper, versionIndex: -1}
	if len(versionField) > 0 {
		index, versionJson, versionBson := FindFieldByName(modelType, versionField)
		if index >= 0 {
			adapter.versionIndex = index
			adapter.versionJson = versionJson
			adapter.versionBson = versionBson
		}
	}
	return adapter
}
func NewRepositoryWithVersion[T any, K any](db *mongo.Database, collectionName string, versionField string, options ...Mapper[T]) *Repository[T, K] {
	return NewMongoRepositoryWithVersion[T, K](db, collectionName, false, versionField, options...)
}
func NewRepository[T any, K any](db *mongo.Database, collectionName string, options ...Mapper[T]) *Repository[T, K] {
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
	if a.Mapper != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			a.Mapper.DbToModel(&objs[i])
		}
	}
	return objs, nil
}
func (a *Repository[T, K]) Load(ctx context.Context, id K) (*T, error) {
	var res T
	if a.ObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return nil, err
		}
		query := bson.M{"_id": objectId}
		ok, er0 := mgo.FindOne(ctx, a.Collection, query, &res)
		if ok && er0 == nil && a.Mapper != nil {
			a.Mapper.DbToModel(&res)
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
	if a.Mapper != nil {
		a.Mapper.DbToModel(&res)
	}
	return &res, er2
}

func (a *Repository[T, K]) Exist(ctx context.Context, id K) (bool, error) {
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
func (a *Repository[T, K]) Create(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	vo := reflect.Indirect(reflect.ValueOf(model))
	if a.versionIndex >= 0 {
		setVersion(vo, a.versionIndex)
	}
	res, rid, err := InsertOne(ctx, a.Collection, model)
	if err != nil {
		return res, err
	}
	if rid != nil {
		idF := vo.Field(a.idIndex)
		switch idF.Kind() {
		case reflect.String:
			idF.Set(reflect.ValueOf(rid.Hex()))
		case reflect.Ptr:
			if idF.Type().String() == "*string" {
				s := rid.Hex()
				idF.Set(reflect.ValueOf(&s))
			} else {
				idF.Set(reflect.ValueOf(rid))
			}
		case reflect.Array:
			idF.Set(reflect.ValueOf(*rid))
		default:
		}
	}
	return res, err
}
func (a *Repository[T, K]) Update(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	vo := reflect.Indirect(reflect.ValueOf(model))
	id := vo.Field(a.idIndex).Interface()
	if a.versionIndex >= 0 {
		currentVersion := vo.Field(a.versionIndex).Interface()
		increaseVersion(vo, a.versionIndex, currentVersion)
		var filter = bson.D{}
		filter = append(filter, bson.E{Key: "_id", Value: id})
		filter = append(filter, bson.E{Key: a.versionBson, Value: currentVersion})
		return UpdateOneByFilter(ctx, a.Collection, filter, model)
	}
	return UpdateOne(ctx, a.Collection, id, model)
}

func (a *Repository[T, K]) Patch(ctx context.Context, model map[string]interface{}) (int64, error) {
	if a.Mapper != nil {
		model = a.Mapper.MapToDb(model)
	}
	id, exist := model[a.idJson]
	if !exist {
		return -1, fmt.Errorf("%s must be in map[string]interface{} for patch", a.idJson)
	}
	if a.versionIndex >= 0 {
		currentVersion, vok := model[a.versionJson]
		if !vok {
			return -1, fmt.Errorf("%s must be in model for patch", a.versionJson)
		}
		ok := increaseMapVersion(model, a.versionJson, currentVersion)
		if !ok {
			return -1, errors.New("Do not support this version type")
		}
		var filter = bson.D{}
		filter = append(filter, bson.E{Key: "_id", Value: id})
		filter = append(filter, bson.E{Key: a.versionBson, Value: currentVersion})
		b := mgo.MapToBson(model, a.Map)
		return PatchOneByFilter(ctx, a.Collection, filter, b)
	}
	b := mgo.MapToBson(model, a.Map)
	return mgo.PatchOne(ctx, a.Collection, id, b)
}

func (a *Repository[T, K]) Save(ctx context.Context, model *T) (int64, error) {
	if a.Mapper != nil {
		a.Mapper.ModelToDb(model)
	}
	vo := reflect.Indirect(reflect.ValueOf(model))
	id := vo.Field(a.idIndex).Interface()
	sid, ok := id.(string)
	if id == nil || ok && len(sid) == 0 {
		if a.versionIndex >= 0 {
			setVersion(vo, a.versionIndex)
		}
		res, rid, err := InsertOne(ctx, a.Collection, model)
		if err != nil {
			return res, err
		}
		if rid != nil {
			idF := vo.Field(a.idIndex)
			switch idF.Kind() {
			case reflect.String:
				idF.Set(reflect.ValueOf(rid.Hex()))
			case reflect.Ptr:
				if idF.Type().String() == "*string" {
					s := rid.Hex()
					idF.Set(reflect.ValueOf(&s))
				} else {
					idF.Set(reflect.ValueOf(rid))
				}
			case reflect.Array:
				idF.Set(reflect.ValueOf(*rid))
			default:
			}
		}
		return res, err
	} else {
		if a.versionIndex >= 0 {
			currentVersion := vo.Field(a.versionIndex).Interface()
			increaseVersion(vo, a.versionIndex, currentVersion)
			var filter = bson.D{}
			filter = append(filter, bson.E{Key: "_id", Value: id})
			filter = append(filter, bson.E{Key: a.versionBson, Value: currentVersion})
			return UpsertOneByFilter(ctx, a.Collection, filter, model)
		} else {
			return UpsertOne(ctx, a.Collection, id, model)
		}
	}
}
func (a *Repository[T, K]) Delete(ctx context.Context, id K) (int64, error) {
	if a.ObjectId {
		objId := fmt.Sprintf("%v", id)
		objectId, err := primitive.ObjectIDFromHex(objId)
		if err != nil {
			return 0, err
		}
		return DeleteOne(ctx, a.Collection, objectId)
	}
	return DeleteOne(ctx, a.Collection, id)
}
func DeleteOne(ctx context.Context, collection *mongo.Collection, id interface{}) (int64, error) {
	filter := bson.M{"_id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if result == nil {
		return 0, err
	}
	return result.DeletedCount, err
}

func setVersion(vo reflect.Value, versionIndex int) bool {
	versionType := vo.Field(versionIndex).Type().String()
	switch versionType {
	case "int32":
		vo.Field(versionIndex).Set(reflect.ValueOf(int32(1)))
		return true
	case "int":
		vo.Field(versionIndex).Set(reflect.ValueOf(1))
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

func increaseVersion(vo reflect.Value, versionIndex int, curVer interface{}) bool {
	versionType := vo.Field(versionIndex).Type().String()
	switch versionType {
	case "int32":
		nextVer := curVer.(int32) + 1
		vo.Field(versionIndex).Set(reflect.ValueOf(nextVer))
		return true
	case "int":
		nextVer := curVer.(int) + 1
		vo.Field(versionIndex).Set(reflect.ValueOf(nextVer))
		return true
	case "int64":
		nextVer := curVer.(int64) + 1
		vo.Field(versionIndex).Set(reflect.ValueOf(nextVer))
		return true
	default:
		return false
	}
}
func UpdateOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model interface{}) (int64, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}
func UpdateOne(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) (int64, error) { //Patch
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}

func increaseMapVersion(model map[string]interface{}, name string, currentVersion interface{}) bool {
	if versionI32, ok := currentVersion.(int32); ok {
		model[name] = versionI32 + 1
		return true
	} else if versionI, ok := currentVersion.(int); ok {
		model[name] = versionI + 1
		return true
	} else if versionI64, ok := currentVersion.(int64); ok {
		model[name] = versionI64 + 1
		return true
	} else {
		return false
	}
}
func PatchOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model map[string]interface{}) (int64, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, filter, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}

func UpsertOne(ctx context.Context, collection *mongo.Collection, id interface{}, model interface{}) (int64, error) {
	filter := bson.M{"_id": id}
	updateQuery := bson.M{
		"$set": model,
	}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, updateQuery, opts)
	if res.ModifiedCount > 0 {
		return res.ModifiedCount, err
	} else if res.UpsertedCount > 0 {
		return res.UpsertedCount, err
	} else {
		return res.MatchedCount, err
	}
}
func UpsertOneByFilter(ctx context.Context, collection *mongo.Collection, filter bson.D, model interface{}) (int64, error) {
	updateQuery := bson.M{
		"$set": model,
	}
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, updateQuery, opts)
	if res.ModifiedCount > 0 {
		return res.ModifiedCount, err
	} else if res.UpsertedCount > 0 {
		return res.UpsertedCount, err
	} else {
		return res.MatchedCount, err
	}
}
