package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"reflect"
	"strings"
)

func SetupMongo(ctx context.Context, c MongoConfig) (*mongo.Database, error) {
	option := options.Client().ApplyURI(c.Uri)
	if c.MaxPoolSize > 0 {
		option = option.SetMaxPoolSize(c.MaxPoolSize)
	}
	if c.MinPoolSize > 0 {
		option = option.SetMinPoolSize(c.MinPoolSize)
	}
	if c.ConnectTimeout > 0 {
		option = option.SetConnectTimeout(c.ConnectTimeout)
	}
	if c.SocketTimeout > 0 {
		option = option.SetSocketTimeout(c.SocketTimeout)
	}
	if c.ServerSelectionTimeout > 0 {
		option = option.SetServerSelectionTimeout(c.ServerSelectionTimeout)
	}
	if c.LocalThreshold > 0 {
		option = option.SetLocalThreshold(c.LocalThreshold)
	}
	if c.HeartbeatInterval > 0 {
		option = option.SetHeartbeatInterval(c.HeartbeatInterval)
	}
	if c.ZlibLevel != 0 {
		option = option.SetZlibLevel(c.ZlibLevel)
	}

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		return nil, err
	}

	db := client.Database(c.Database)
	return db, nil
}

func CreateConnection(ctx context.Context, uri string, database string) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(database)
	return db, nil
}

func CreateUniqueIndex(collection *mongo.Collection, fieldName string) (string, error) {
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bsonx.Doc{{fieldName, bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		},
	)
	return indexName, err
}

func FindOneWithId(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool, modelType reflect.Type) (interface{}, error) {
	if objectId {
		objId := id.(string)
		return FindOneWithObjectId(ctx, collection, objId, modelType)
	}
	return FindOne(ctx, collection, bson.M{"_id": id}, modelType)
}

func FindOneWithObjectId(ctx context.Context, collection *mongo.Collection, id string, modelType reflect.Type) (interface{}, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return FindOne(ctx, collection, bson.M{"_id": objectId}, modelType)
}

func FindOne(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return nil, nil
		}
		return nil, x.Err()
	}
	result := reflect.New(modelType).Interface()
	er2 := x.Decode(result)
	if er2 != nil {
		if strings.Contains(fmt.Sprint(er2), "cannot decode") {
			return result, nil
		}
		return nil, er2
	}
	return result, nil
}

func FindOneAndDecode(ctx context.Context, collection *mongo.Collection, query bson.M, result interface{}) (bool, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return false, nil
		}
		return false, x.Err()
	}
	er2 := x.Decode(result)
	return true, er2
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return diff
}

func FindByIds(ctx context.Context, collection *mongo.Collection, ids []string, idObjectId bool, modelType reflect.Type) (interface{}, []string, error) {
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	result := reflect.New(modelsType).Interface()
	res := reflect.Indirect(reflect.ValueOf(result))
	if !idObjectId {
		find, errFind := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if errFind != nil {
			return nil, ids, errFind
		}
		err := find.All(ctx, result)
		keySuccess := []string{""}
		_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			key := res.Index(i).FieldByName(fieldName).String()
			keySuccess = append(keySuccess, key)
		}
		keys := difference(keySuccess, ids)
		_ = find.Close(ctx)
		return result, keys, err
	} else {
		id := make([]primitive.ObjectID, 0)
		for _, val := range ids {
			item, err := primitive.ObjectIDFromHex(val)
			if err != nil {
				return false, nil, err
			}
			id = append(id, item)
		}
		find, errFind := collection.Find(ctx, bson.M{"_id": bson.M{"$in": id}})
		if errFind != nil {
			return false, ids, errFind
		}
		err := find.All(ctx, result)
		keySuccess := make([]primitive.ObjectID, 0)
		_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			key, _ := primitive.ObjectIDFromHex(res.Index(i).FieldByName(fieldName).Interface().(primitive.ObjectID).Hex())
			keySuccess = append(keySuccess, key)
		}
		keyToStr := []string{""}
		for index, _ := range keySuccess {
			key := keySuccess[index].Hex()
			keyToStr = append(keyToStr, key)
		}
		keys := difference(keyToStr, ids)
		return result, keys, err
	}
}

func FindByIdsAndDecode(ctx context.Context, collection *mongo.Collection, ids []string, idObjectId bool, result interface{}) ([]string, error) {
	var keys []string
	if !idObjectId {
		res := reflect.Indirect(reflect.ValueOf(result))
		find, errFind := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if errFind != nil {
			return ids, errFind
		}
		find.All(ctx, result)
		keySuccess := []string{""}
		//		_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			keySuccess = append(keySuccess, res.Index(i).Field(0).String())
		}
		keys = difference(keySuccess, ids)
		_ = find.Close(ctx)
	} else {
		res := reflect.Indirect(reflect.ValueOf(result))
		id := make([]primitive.ObjectID, 0)
		for _, val := range ids {
			item, err := primitive.ObjectIDFromHex(val)
			if err != nil {
				return nil, err
			}
			id = append(id, item)
		}
		find, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": id}})
		if err != nil {
			return ids, err
		}
		find.All(ctx, result)
		keySuccess := make([]primitive.ObjectID, 0)
		//_, fieldName := FindIdField(modelType)
		for i := 0; i < res.Len(); i++ {
			key, _ := primitive.ObjectIDFromHex(res.Index(i).Field(0).Interface().(primitive.ObjectID).Hex())
			keySuccess = append(keySuccess, key)
		}
		keyToStr := []string{""}
		for index, _ := range keySuccess {
			key := keySuccess[index].Hex()
			keyToStr = append(keyToStr, key)
		}
		keys = difference(keyToStr, ids)
	}
	if result != nil {
		return keys, nil
	}
	return keys, errors.New("no result return")
}

func Find(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	arr := reflect.New(modelsType).Interface()
	er2 := cur.All(ctx, arr)
	_ = cur.Close(ctx)
	return arr, er2
}

func FindAndDecode(ctx context.Context, collection *mongo.Collection, query bson.M, arr interface{}) (bool, error) {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return false, err
	}
	er2 := cur.All(ctx, arr)
	return true, er2
}

func Exist(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool) (bool, error) {
	query := bson.M{"_id": id}
	if objectId {
		objId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			return false, err
		}
		query = bson.M{"_id": objId}
	}
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if fmt.Sprint(x.Err()) == "mongo: no documents in result" {
			return false, nil
		} else {
			return false, x.Err()
		}
	}
	return true, nil
}

func DeleteOne(ctx context.Context, coll *mongo.Collection, query bson.M) (int64, error) {
	result, err := coll.DeleteOne(ctx, query)
	return result.DeletedCount, err
}

func InsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (int64, error) {
	result, err := collection.InsertOne(ctx, model)
	if err != nil {
		errMsg := err.Error()
		if strings.Index(errMsg, "duplicate key error collection:") >= 0 {
			if strings.Index(errMsg, "dup key: { _id: ") >= 0 {
				return -1, nil
			} else {
				return -2, nil
			}
		} else {
			return 0, err
		}
	} else {
		if idValue, ok := result.InsertedID.(primitive.ObjectID); ok {
			valueOfModel := reflect.Indirect(reflect.ValueOf(model))
			typeOfModel := valueOfModel.Type()
			idIndex, _ := FindIdField(typeOfModel)
			if idIndex != -1 {
				mapObjectIdToModel(idValue, valueOfModel, idIndex)
			}
		}
		return 1, err
	}
}

func InsertOneWithVersion(ctx context.Context, collection *mongo.Collection, model interface{}, versionIndex int) (int64, error) {
	var defaultVersion interface{}
	modelType := reflect.TypeOf(model).Elem()
	versionType := modelType.Field(versionIndex).Type
	switch versionType.String() {
	case "int":
		defaultVersion = int(1)
	case "int32":
		defaultVersion = int32(1)
	case "int64":
		defaultVersion = int64(1)
	default:
		panic("not support type's version")
	}
	model, err := setValue(model, versionIndex, defaultVersion)
	if err != nil {
		return 0, err
	}
	return InsertOne(ctx, collection, model)
}

func InsertMany(ctx context.Context, collection *mongo.Collection, models interface{}) (bool, error) {
	arr := make([]interface{}, 0)
	values := reflect.Indirect(reflect.ValueOf(models))
	length := values.Len()
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		for i := 0; i < length; i++ {
			arr = append(arr, values.Index(i).Interface())
		}
	}

	if len(arr) > 0 {
		res, err := collection.InsertMany(ctx, arr)
		if err != nil {
			if strings.Index(err.Error(), "duplicate key error collection:") >= 0 {
				return true, nil
			} else {
				return false, err
			}
		}

		valueOfModel := reflect.Indirect(reflect.ValueOf(arr[0]))
		idIndex, _ := FindIdField(valueOfModel.Type())
		if idIndex >= 0 {
			for i, _ := range arr {
				if idValue, ok := res.InsertedIDs[i].(primitive.ObjectID); ok {
					mapObjectIdToModel(idValue, values.Index(i), idIndex)
				}
			}
		}
	}
	return false, nil
}

//For Insert
func mapObjectIdToModel(id primitive.ObjectID, valueOfModel reflect.Value, idIndex int) {
	switch reflect.Indirect(valueOfModel).Field(idIndex).Kind() {
	case reflect.String:
		if _, err := setValue(valueOfModel, idIndex, id.Hex()); err != nil {
			log.Println("Err: ", err)
		}
		break
	default:
		if _, err := setValue(valueOfModel, idIndex, id); err != nil {
			log.Println("Err: ", err)
		}
		break
	}
}

func InsertManySkipErrors(ctx context.Context, collection *mongo.Collection, models interface{}) (interface{}, interface{}, error) {
	arr := make([]interface{}, 0)
	indexFailArr := make([]int, 0)
	modelsType := reflect.TypeOf(models)
	insertedFails := reflect.New(modelsType).Interface()
	idName := ""
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		if values.Len() == 0 {
			return insertedFails, insertedFails, nil
		}
		_, name := FindIdField(reflect.TypeOf(values.Index(0).Interface()))
		idName = name
		for i := 0; i < values.Len(); i++ {
			arr = append(arr, values.Index(i).Interface())
		}
	}
	var defaultOrdered = false
	rs, err := collection.InsertMany(ctx, arr, &options.InsertManyOptions{Ordered: &defaultOrdered})
	if err != nil {
		values := reflect.ValueOf(models)
		insertedSuccess := reflect.New(modelsType).Interface()
		if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
			for _, writeError := range bulkWriteException.WriteErrors {
				appendToArray(insertedFails, values.Index(writeError.Index).Interface())
				indexFailArr = append(indexFailArr, writeError.Index)
			}
			if rs != nil && len(idName) > 0 {
				insertedSuccess = mapIdInObjects(models, indexFailArr, rs.InsertedIDs, modelsType, idName)
			}
			return insertedSuccess, insertedFails, err
		} else {
			for i := 0; i < values.Len(); i++ {
				appendToArray(insertedFails, values.Index(i).Interface())
			}
			return insertedSuccess, insertedFails, err
		}
	}
	if len(idName) > 0 {
		insertedSuccess := mapIdInObjects(models, indexFailArr, rs.InsertedIDs, modelsType, idName)
		return insertedSuccess, nil, err
	}
	return nil, nil, err
}

func mapIdInObjects(models interface{}, arrayFailIndexIgnore []int, insertedIDs []interface{}, modelsType reflect.Type, fieldName string) interface{} {
	insertedSuccess := reflect.New(modelsType).Interface()
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		if length := values.Len(); length > 0 {
			if index := findIndex(values.Index(0).Interface(), fieldName); index != -1 {
				for i := 0; i < length; i++ {
					if !existInArray(arrayFailIndexIgnore, i) {
						if id, ok := insertedIDs[i].(primitive.ObjectID); ok {
							itemValue := values.Index(i)
							var errSet error
							var vSet interface{}
							switch reflect.Indirect(itemValue).FieldByName(fieldName).Kind() {
							case reflect.String:
								idString := id.Hex()
								vSet, errSet = setValue(itemValue, index, idString)
								break
							default:
								vSet, errSet = setValue(itemValue, index, id)
								break
							}
							if errSet == nil {
								appendToArray(insertedSuccess, vSet)
							} else {
								appendToArray(insertedSuccess, itemValue.Interface())
								log.Println("Error map Id: ", errSet)
							}
						}
					}
				}
			}
		}
	}
	return insertedSuccess
}

func existInArray(arr []int, value interface{}) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func UpdateOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (int64, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, query, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}

func UpdateMany(ctx context.Context, collection *mongo.Collection, models interface{}, filter interface{}) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		length := values.Len()
		if length > 0 {
			modelType := reflect.Indirect(reflect.ValueOf(values.Index(0).Interface())).Type()
			if index, _ := FindIdField(modelType); index != -1 {
				for i := 0; i < length; i++ {
					row := values.Index(i).Interface()
					updateModel := mongo.NewUpdateOneModel().SetUpdate(row).SetFilter(filter)
					if filter == nil {
						v, _ := getValue(row, index)
						updateModel = mongo.NewUpdateOneModel().SetUpdate(row).SetFilter(bson.M{"_id": v})
					}
					models_ = append(models_, updateModel)
				}
			}
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}

func UpdateMaps(ctx context.Context, collection *mongo.Collection, maps []map[string]interface{}, idName string) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	for _, row:= range maps{
		v, _ := row[idName]
		if v!=nil{
			updateModel := mongo.NewUpdateOneModel().SetUpdate(bson.M{
				"$set": row,
			}).SetFilter(bson.M{"_id": v})
			models_ = append(models_, updateModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}

func UpsertMaps(ctx context.Context, collection *mongo.Collection, maps []map[string]interface{}, idName string) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	for _, row:= range maps{
		id, _ := row[idName]
		if id != nil || (reflect.TypeOf(id).String() == "string") || (reflect.TypeOf(id).String() == "string" && len(id.(string)) > 0) {
			updateModel := mongo.NewUpdateOneModel().SetUpdate(bson.M{
				"$set": row,
			}).SetUpsert(true).SetFilter(bson.M{"_id": id})
			models_ = append(models_, updateModel)
		}else{
			insertModel := mongo.NewInsertOneModel().SetDocument(row)
			models_ = append(models_, insertModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}

func PatchOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (int64, error) {
	updateQuery := bson.M{
		"$set": model,
	}
	result, err := collection.UpdateOne(ctx, query, updateQuery)
	if result.ModifiedCount > 0 {
		return result.ModifiedCount, err
	} else if result.UpsertedCount > 0 {
		return result.UpsertedCount, err
	} else {
		return result.MatchedCount, err
	}
}

func PatchMany(ctx context.Context, collection *mongo.Collection, models interface{}, idName string) (*mongo.BulkWriteResult, error) {
	models_ := make([]mongo.WriteModel, 0)
	ids := make([]interface{}, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		length := values.Len()
		if length > 0 {
			if index := findIndex(values.Index(0).Interface(), idName); index != -1 {
				for i := 0; i < length; i++ {
					row := values.Index(i).Interface()
					updateModel := mongo.NewUpdateOneModel().SetUpdate(values.Index(i))
					v, err1 := getValue(row, index)
					if err1 == nil && v != nil {
						if reflect.TypeOf(v).String() != "string" {
							updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
								"$set": row,
							}).SetFilter(bson.M{"_id": v})
						} else {
							if idStr, ok := v.(string); ok {
								updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
									"$set": row,
								}).SetFilter(bson.M{"_id": idStr})
							}
						}
						ids = append(ids, v)
					}
					models_ = append(models_, updateModel)
				}
			}
		}
	}
	var defaultOrdered = false
	return collection.BulkWrite(ctx, models_, &options.BulkWriteOptions{Ordered: &defaultOrdered})
}

func PatchManyAndGetSuccessList(ctx context.Context, collection *mongo.Collection, models interface{}, idName string) (interface{}, interface{}, []interface{}, error) {
	models_ := make([]mongo.WriteModel, 0)
	ids := make([]interface{}, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		length := values.Len()
		if length > 0 {
			if index := findIndex(values.Index(0).Interface(), idName); index != -1 {
				for i := 0; i < length; i++ {
					row := values.Index(i).Interface()
					updateModel := mongo.NewUpdateOneModel().SetUpdate(values.Index(i))
					v, err1 := getValue(row, index)
					if err1 == nil && v != nil {
						if reflect.TypeOf(v).String() != "string" {
							updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
								"$set": row,
							}).SetFilter(bson.M{"_id": v})
						} else {
							if idStr, ok := v.(string); ok {
								updateModel = mongo.NewUpdateOneModel().SetUpdate(bson.M{
									"$set": row,
								}).SetFilter(bson.M{"_id": idStr})
							}
						}
						ids = append(ids, v)
					}
					models_ = append(models_, updateModel)
				}
			}
		}
	}
	var defaultOrdered = false
	_, er0 := collection.BulkWrite(ctx, models_, &options.BulkWriteOptions{Ordered: &defaultOrdered})
	if er0 != nil {
		return nil, nil, nil, er0
	}
	successIdList := make([]interface{}, 0)
	_options := options.FindOptions{Projection: bson.M{"_id": 1}}
	cur, er1 := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, &_options)
	if er1 != nil {
		return nil, nil, nil, er1
	}
	er2 := cur.All(ctx, &successIdList)
	if er2 != nil {
		return nil, nil, nil, er2
	}
	successIdList = mapArrayInterface(successIdList)
	failList, failIdList := diffModelArray(models, successIdList, idName)
	return successIdList, failList, failIdList, er0
}

func mapArrayInterface(successIdList []interface{}) []interface{} {
	arr := make([]interface{}, 0)
	for _, value := range successIdList {
		if primitiveE, ok := value.(primitive.D); ok {
			for _, itemPrimitiveE := range primitiveE {
				arr = append(arr, itemPrimitiveE.Value)
			}
		}
	}
	return arr
}

func diffModelArray(modelsAll interface{}, successIdList interface{}, idName string) (interface{}, []interface{}) {
	modelsB := make([]interface{}, 0)
	modelBId := make([]interface{}, 0)
	switch reflect.TypeOf(modelsAll).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(modelsAll)
		length := values.Len()
		if length > 0 {
			if index := findIndex(values.Index(0).Interface(), idName); index != -1 {
				for i := 0; i < length; i++ {
					itemValue := values.Index(i)
					id, _ := getValue(itemValue.Interface(), index)
					if !existInArrayInterface(successIdList, id) {
						modelsB = append(modelsB, itemValue.Interface())
						modelBId = append(modelBId, id)
					}
				}
			}
		}
	}
	return modelsB, modelBId
}

func existInArrayInterface(arr interface{}, valueID interface{}) bool {
	switch reflect.TypeOf(arr).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(arr)
		for i := 0; i < values.Len(); i++ {
			itemValueID := values.Index(i).Interface()
			if itemValueID == valueID {
				return true
			}
		}
	}
	return false
}

func UpsertOne(ctx context.Context, collection *mongo.Collection, filter bson.M, model interface{}) (int64, error) {
	defaultObjID, _ := primitive.ObjectIDFromHex("000000000000")

	if idValue := filter["_id"]; idValue == "" || idValue == 0 || idValue == defaultObjID {
		return InsertOne(ctx, collection, model)
	} else {
		isExisted, err := Exist(ctx, collection, idValue, false)
		if err != nil {
			return 0, err
		}
		if isExisted {
			update := bson.M{
				"$set": model,
			}
			result := collection.FindOneAndUpdate(ctx, filter, update)
			if result.Err() != nil {
				if fmt.Sprint(result.Err()) == "mongo: no documents in result" {
					return 0, nil
				} else {
					return 0, result.Err()
				}
			}
			return 1, result.Err()
		} else {
			return InsertOne(ctx, collection, model)
		}
	}
}

func UpsertOneWithVersion(ctx context.Context, collection *mongo.Collection, model interface{}, versionIndex int) (int64, error) {
	idQuery := BuildQueryByIdFromObject(model)
	defaultObjID, _ := primitive.ObjectIDFromHex("000000000000")

	if idValue := idQuery["_id"]; idValue == "" || idValue == 0 || idValue == defaultObjID {
		return InsertOneWithVersion(ctx, collection, model, versionIndex)
	} else {
		isExisted, err := Exist(ctx, collection, idValue, false)
		if err != nil {
			return 0, err
		}
		if isExisted {
			versionQuery := BuildIdAndVersionQueryByVersionIndex(idQuery, model, versionIndex)
			update := bson.M{
				"$set": model,
			}
			result := collection.FindOneAndUpdate(ctx, versionQuery, update)
			if result.Err() != nil {
				if fmt.Sprint(result.Err()) == "mongo: no documents in result" {
					return -1, nil
				} else {
					return 0, result.Err()
				}
			}
			return 1, result.Err()
		} else {
			return InsertOneWithVersion(ctx, collection, model, versionIndex)
		}
	}
}

func UpsertMany(ctx context.Context, collection *mongo.Collection, model interface{}) (*mongo.BulkWriteResult, error) { //Patch
	models := make([]mongo.WriteModel, 0)
	switch reflect.TypeOf(model).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(model)

		n := values.Len()
		for i := 0; i < n; i++ {
			row := values.Index(i).Interface()
			if v, ok := row.(bson.M); ok {
				id := v["_id"]
				if id != nil || (reflect.TypeOf(id).String() == "string") || (reflect.TypeOf(id).String() == "string" && len(id.(string)) > 0) { // if exist
					updateModel := mongo.NewReplaceOneModel().SetUpsert(true).SetReplacement(row).SetFilter(bson.M{"_id": id})
					models = append(models, updateModel)
				} else {
					insertModel := mongo.NewInsertOneModel().SetDocument(row)
					models = append(models, insertModel)
				}
			}
		}
	}
	rs, err := collection.BulkWrite(ctx, models)
	return rs, err
}

//For Get By Id
func FindFieldIndex(modelType reflect.Type, fieldName string) int {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		if field.Name == fieldName {
			return i
		}
	}
	return -1
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

func FindIdField(modelType reflect.Type) (int, string) {
	return FindField(modelType, "_id")
}

func FindField(modelType reflect.Type, bsonName string) (int, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		bsonTag := field.Tag.Get("bson")
		tags := strings.Split(bsonTag, ",")
		for _, tag := range tags {
			if strings.TrimSpace(tag) == bsonName {
				return i, field.Name
			}
		}
	}
	return -1, ""
}

func GetFieldByJson(modelType reflect.Type, jsonName string) (int, string, string) {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		tag1, ok1 := field.Tag.Lookup("json")
		if ok1 && strings.Split(tag1, ",")[0] == jsonName {
			if tag2, ok2 := field.Tag.Lookup("bson"); ok2 {
				return i, field.Name, strings.Split(tag2, ",")[0]
			}
			return i, field.Name, ""
		}
	}
	return -1, jsonName, jsonName
}

//For Search and Patch
func GetBsonName(modelType reflect.Type, fieldName string) string {
	field, found := modelType.FieldByName(fieldName)
	if !found {
		return fieldName
	}
	if tag, ok := field.Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return fieldName
}

func GetBsonNameByIndex(model interface{}, fieldIndex int) string {
	modelType := reflect.TypeOf(model).Elem()
	if tag, ok := modelType.Field(fieldIndex).Tag.Lookup("bson"); ok {
		return strings.Split(tag, ",")[0]
	}
	return ""
}

//For Update
func BuildQueryByIdFromObject(object interface{}) bson.M {
	valueOf := reflect.Indirect(reflect.ValueOf(object))
	if idIndex, _ := FindIdField(valueOf.Type()); idIndex != -1 {
		value := valueOf.Field(idIndex).Interface()
		return bson.M{"_id": value}
	} else {
		panic("id field not found")
	}
}

//For Patch
func BuildQueryByIdFromMap(m map[string]interface{}, idName string) bson.M {
	if idValue, exist := m[idName]; exist {
		return bson.M{"_id": idValue}
	} else {
		panic("id field not found")
	}
}

func MapToBson(object map[string]interface{}, objectMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range object {
		field := objectMap[key]
		result[field] = value
	}
	return result
}

func MakeMapBson(modelType reflect.Type) map[string]string {
	maps := make(map[string]string)
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		key1 := modelType.Field(i).Name
		fields, _ := modelType.FieldByName(key1)
		if tag, ok := fields.Tag.Lookup("bson"); ok {
			if strings.Contains(tag, ",") {
				a := strings.Split(tag, ",")
				maps[key1] = a[0]
			} else {
				maps[key1] = tag
			}
		} else {
			maps[key1] = key1
		}
	}
	return maps
}

// For Batch Update
func initArrayResults(modelsType reflect.Type) interface{} {
	return reflect.New(modelsType).Interface()
}

func appendToArray(arr interface{}, item interface{}) interface{} {
	arrValue := reflect.ValueOf(arr)
	elemValue := reflect.Indirect(arrValue)

	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = reflect.Indirect(itemValue)
	}
	elemValue.Set(reflect.Append(elemValue, itemValue))
	return arr
}

func MapToMongoObjects(model interface{}, idName string, idObjectId bool, modelType reflect.Type, newId bool) (interface{}, interface{}) {
	var results = initArrayResults(modelType)
	var ids = make([]interface{}, 0)
	switch reflect.TypeOf(model).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(model)
		for i := 0; i < values.Len(); i++ {
			model, id := MapToMongoObject(values.Index(i).Interface(), idName, idObjectId, newId)
			ids = append(ids, id)
			results = appendToArray(results, model)
		}
	}
	return results, ids
}

func MapToMongoObject(model interface{}, idName string, objectId bool, newId bool) (interface{}, interface{}) {
	if index := findIndex(model, idName); index != -1 {
		id, _ := getValue(model, index)
		if objectId {
			if newId && (id == nil) {
				setValue(model, index, bsonx.ObjectID(primitive.NewObjectID()))
			} else {
				objectId, err := primitive.ObjectIDFromHex(id.(string))
				if err == nil {
					setValue(model, index, objectId)
				}
			}
		} else {
			setValue(model, index, id)
		}
		return model, id
	}
	return model, nil
}

func getValue(model interface{}, index int) (interface{}, error) {
	valueObject := reflect.Indirect(reflect.ValueOf(model))
	return valueObject.Field(index).Interface(), nil
}

func setValue(model interface{}, index int, value interface{}) (interface{}, error) {
	valueObject := reflect.Indirect(reflect.ValueOf(model))
	switch reflect.ValueOf(model).Kind() {
	case reflect.Ptr:
		{
			valueObject.Field(index).Set(reflect.ValueOf(value))
			return model, nil
		}
	default:
		if modelWithTypeValue, ok := model.(reflect.Value); ok {
			_, err := setValueWithTypeValue(modelWithTypeValue, index, value)
			return modelWithTypeValue.Interface(), err
		}
	}
	return model, nil
}

func setValueWithTypeValue(model reflect.Value, index int, value interface{}) (reflect.Value, error) {
	trueValue := reflect.Indirect(model)
	switch trueValue.Kind() {
	case reflect.Struct:
		{
			val := reflect.Indirect(reflect.ValueOf(value))
			if trueValue.Field(index).Kind() == val.Kind() {
				trueValue.Field(index).Set(reflect.ValueOf(value))
				return trueValue, nil
			} else {
				return trueValue, fmt.Errorf("value's kind must same as field's kind")
			}
		}
	default:
		return trueValue, nil
	}
}

func findIndex(model interface{}, fieldName string) int {
	modelType := reflect.Indirect(reflect.ValueOf(model))
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		if modelType.Type().Field(i).Name == fieldName {
			return i
		}
	}
	return -1
}

//Version
func copyMap(originalMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range originalMap {
		newMap[k] = v
	}
	return newMap
}

func BuildIdAndVersionQueryByMap(query map[string]interface{}, v map[string]interface{}, maps map[string]string, versionField string) map[string]interface{} {
	newMap := copyMap(query)
	if currentVersion, exist := v[versionField]; exist {
		newMap[maps[versionField]] = currentVersion
		switch versionValue := currentVersion.(type) {
		case int:
			{
				v[versionField] = versionValue + 1
			}
		case int32:
			{
				v[versionField] = versionValue + 1
			}
		case int64:
			{
				v[versionField] = versionValue + 1
			}
		default:
			panic("not support type's version")
		}
	}
	return newMap
}

func BuildIdAndVersionQuery(query map[string]interface{}, model interface{}, versionField string) map[string]interface{} {
	index := findIndex(model, versionField)
	return BuildIdAndVersionQueryByVersionIndex(query, model, index)
}

func BuildIdAndVersionQueryByVersionIndex(query map[string]interface{}, model interface{}, versionIndex int) map[string]interface{} {
	newMap := copyMap(query)
	valueOfModel := reflect.Indirect(reflect.ValueOf(model))
	if versionIndex >= 0 && versionIndex < valueOfModel.NumField() {
		var valueOfCurrentVersion reflect.Value
		valueOfCurrentVersion = valueOfModel.Field(versionIndex)
		versionColumnName := GetBsonNameByIndex(model, versionIndex)
		newMap[versionColumnName] = valueOfCurrentVersion.Interface()
		switch valueOfCurrentVersion.Kind().String() {
		case "int":
			{
				nextVersion := reflect.ValueOf(valueOfCurrentVersion.Interface().(int) + 1)
				valueOfModel.Field(versionIndex).Set(nextVersion)
			}
		case "int32":
			{
				nextVersion := reflect.ValueOf(valueOfCurrentVersion.Interface().(int32) + 1)
				valueOfModel.Field(versionIndex).Set(nextVersion)
			}
		case "int64":
			{
				nextVersion := reflect.ValueOf(valueOfCurrentVersion.Interface().(int64) + 1)
				valueOfModel.Field(versionIndex).Set(nextVersion)
			}
		default:
			panic("not support type's version")
		}
		return newMap
	} else {
		panic("invalid versionIndex")
	}
}

func UpdateByIdAndVersion(ctx context.Context, collection *mongo.Collection, model interface{}, versionIndex int) (int64, error) {
	idQuery := BuildQueryByIdFromObject(model)
	versionQuery := BuildIdAndVersionQueryByVersionIndex(idQuery, model, versionIndex)
	rowAffect, er1 := UpdateOne(ctx, collection, model, versionQuery)
	if er1 != nil {
		return 0, er1
	}
	if rowAffect == 0 {
		isExist, er2 := Exist(ctx, collection, idQuery["_id"], false)
		if er2 != nil {
			return 0, er2
		}
		if isExist {
			return -1, nil
		} else {
			return 0, nil
		}
	}
	return rowAffect, er1
}

func PatchByIdAndVersion(ctx context.Context, collection *mongo.Collection, model map[string]interface{}, maps map[string]string, idName string, versionField string) (int64, error) {
	idQuery := BuildQueryByIdFromMap(model, idName)
	versionQuery := BuildIdAndVersionQueryByMap(idQuery, model, maps, versionField)
	rowAffect, er1 := PatchOne(ctx, collection, MapToBson(model, maps), versionQuery)
	if er1 != nil {
		return 0, er1
	}
	if rowAffect == 0 {
		isExist, er2 := Exist(ctx, collection, idQuery["_id"], false)
		if er2 != nil {
			return 0, er2
		}
		if isExist {
			return -1, nil
		}
		return 0, nil
	}
	return rowAffect, er1
}

func InArray(value int, arr []int) bool {
	for i := 0; i < len(arr); i++ {
		if value == arr[i] {
			return true
		}
	}
	return false
}