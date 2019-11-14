package mongo

import (
	"context"
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

func SetupMongo(ctx context.Context, mongoConfig MongoConfig) (*mongo.Database, error) {
	return CreateConnection(ctx, mongoConfig.Uri, mongoConfig.Database)
}

func CreateConnection(ctx context.Context, uri string, database string) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database(database)
	return db, nil
}

func FindOneWithId(ctx context.Context, collection *mongo.Collection, id interface{}, objectId bool, modelType reflect.Type) (interface{}, error) {
	if objectId {
		objId := id.(string)
		return FindOneWithObjectId(ctx, collection, objId, modelType)
	} else {
		return FindOne(ctx, collection, bson.M{"_id": id}, modelType)
	}
}

func FindOneWithObjectId(ctx context.Context, collection *mongo.Collection, id string, modelType reflect.Type) (interface{}, error) {
	objectId, _ := primitive.ObjectIDFromHex(id)
	return FindOne(ctx, collection, bson.M{"_id": objectId}, modelType)
}

func FindOne(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	x := collection.FindOne(ctx, query)
	if x.Err() != nil {
		if strings.Compare(fmt.Sprint(x.Err()), "mongo: no documents in result") == 0 {
			return nil, nil
		} else {
			return nil, x.Err()
		}
	} else {
		result := reflect.New(modelType).Interface()
		er2 := x.Decode(result)
		if er2 != nil {
			if strings.Contains(fmt.Sprint(er2), "cannot decode null into") {
				return result, nil
			}
			return nil, er2
		} else {
			return result, nil
		}
	}
}

func Find(ctx context.Context, collection *mongo.Collection, query bson.M, modelType reflect.Type) (interface{}, error) {
	cur, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	modelsType := reflect.Zero(reflect.SliceOf(modelType)).Type()
	arr := reflect.New(modelsType).Interface()
	er2 := cur.All(ctx, arr)
	return arr, er2
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
		if strings.Compare(fmt.Sprint(x.Err()), "mongo: no documents in result") == 0 {
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

func InsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (interface{}, bool, error) {
	result, err := collection.InsertOne(ctx, model)
	if err != nil {
		if strings.Index(err.Error(), "duplicate key error collection:") >= 0 {
			return nil, true, nil
		} else {
			return nil, false, err
		}
	} else {
		if idValue, ok := result.InsertedID.(primitive.ObjectID); ok {
			typeOfId := reflect.Indirect(reflect.ValueOf(model)).Type()
			idField := FindIdField(typeOfId)
			mapObjectIdToModel(idValue, model, idField)
		}
		return model, false, err
	}
}

func InsertMany(ctx context.Context, collection *mongo.Collection, models interface{}) (interface{}, bool, error) {
	arr := make([]interface{}, 0)
	values := reflect.ValueOf(models)
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
				return nil, true, nil
			} else {
				return nil, false, err
			}
		}

		var idField string
		valueOfModel := reflect.Indirect(reflect.ValueOf(arr[0]))
		idField = FindIdField(valueOfModel.Type())
		for i, v := range arr {
			if idValue, ok := res.InsertedIDs[i].(primitive.ObjectID); ok {
				mapObjectIdToModel(idValue, v, idField)
			}
		}
	}
	return models, false, nil
}

//For Insert
func mapObjectIdToModel(id primitive.ObjectID, model interface{}, idField string) {
	valueOfModel := reflect.Indirect(reflect.ValueOf(model))
	switch valueOfModel.FieldByName(idField).Kind() {
	case reflect.String:
		valueOfModel.FieldByName(idField).Set(reflect.ValueOf(id.Hex()))
		break
	default:
		valueOfModel.FieldByName(idField).Set(reflect.ValueOf(id))
		break
	}
}

func InsertManySkipErrors(ctx context.Context, collection *mongo.Collection, models interface{}, modelsType reflect.Type, fieldName string) (interface{}, interface{}, error) {
	arr := make([]interface{}, 0)
	indexFailArr := make([]int, 0)
	insertedFails := reflect.New(modelsType).Interface()
	insertedSuccess := reflect.New(modelsType).Interface()
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		for i := 0; i < values.Len(); i++ {
			arr = append(arr, values.Index(i).Interface())
		}
	}
	var defaultOrdered = false
	rs, err := collection.InsertMany(ctx, arr, &options.InsertManyOptions{Ordered: &defaultOrdered})
	if err != nil {
		values := reflect.ValueOf(models)
		if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
			for _, writeError := range bulkWriteException.WriteErrors {
				appendToArray(insertedFails, values.Index(writeError.Index).Interface())
				indexFailArr = append(indexFailArr, writeError.Index)
			}
		}
		insertedSuccess = mapIdInObjects(models, indexFailArr, rs.InsertedIDs, modelsType, fieldName)
		return insertedSuccess, insertedFails, err
	}
	insertedSuccess = mapIdInObjects(models, indexFailArr, rs.InsertedIDs, modelsType, fieldName)
	return insertedSuccess, nil, err
}

func mapIdInObjects(models interface{}, arrayFailIndexIgnore []int, insertedIDs []interface{}, modelsType reflect.Type, fieldName string) interface{} {
	insertedSuccess := reflect.New(modelsType).Interface()
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		k := 0
		for i := 0; i < values.Len(); i++ {
			itemValue := values.Index(i)
			if !existInArray(arrayFailIndexIgnore, i) {
				id := insertedIDs[i]
				_, errSet := setValue(itemValue, fieldName, id)
				if errSet == nil {
					insertedSuccess = appendToArray(insertedSuccess, itemValue.Interface())
					k++
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

func UpdateOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (interface{}, error) { //Patch
	updateQuery := bson.M{
		"$set": model,
	}
	_, err := collection.UpdateOne(ctx, query, updateQuery)
	return model, err
}

func UpdateMany(ctx context.Context, collection *mongo.Collection, models interface{}, filter interface{}) (interface{}, error) {
	models_ := make([]mongo.WriteModel, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		for i := 0; i < values.Len(); i++ {
			row := values.Index(i).Interface()
			updateModel := mongo.NewUpdateOneModel().SetUpdate(row).SetFilter(filter)
			if filter == nil {
				v, _ := getValue(row, "Id")
				updateModel = mongo.NewUpdateOneModel().SetUpdate(row).SetFilter(bson.M{"_id": v})
			}
			models_ = append(models_, updateModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models_)
	return res, err
}

func PatchOne(ctx context.Context, collection *mongo.Collection, model interface{}, query bson.M) (interface{}, error) {
	updateQuery := bson.M{
		"$set": model,
	}
	_, err := collection.UpdateOne(ctx, query, updateQuery)
	return model, err
}

func PatchMany(ctx context.Context, collection *mongo.Collection, models interface{}, idName string) (interface{}, interface{}, []interface{}, error) {
	models_ := make([]mongo.WriteModel, 0)
	ids := make([]interface{}, 0)
	switch reflect.TypeOf(models).Kind() {
	case reflect.Slice:
		values := reflect.ValueOf(models)
		for i := 0; i < values.Len(); i++ {
			row := values.Index(i).Interface()
			updateModel := mongo.NewUpdateOneModel().SetUpdate(values.Index(i))
			v, err1 := getValue(row, idName)
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
	var defaultOrdered = false
	_, err := collection.BulkWrite(ctx, models_, &options.BulkWriteOptions{Ordered: &defaultOrdered})
	successIdList := make([]interface{}, 0)
	_options := options.FindOptions{Projection: bson.M{"_id": 1}}
	cur, errFind := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, &_options)
	if errFind != nil {
		log.Println(errFind)
	} else {
		err1 := cur.All(ctx, &successIdList)
		if err1 == nil {
			successIdList = mapArrayInterface(successIdList)
		}
	}
	failList, failIdList := diffModelArray(models, successIdList, idName)
	return successIdList, failList, failIdList, err
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
		for i := 0; i < values.Len(); i++ {
			itemValue := values.Index(i)
			id, _ := getValue(itemValue.Interface(), idName)
			if !existInArrayInterface(successIdList, id) {
				modelsB = append(modelsB, itemValue.Interface())
				modelBId = append(modelBId, id)
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

func UpsertOne(ctx context.Context, collection *mongo.Collection, model interface{}) (interface{}, error) {
	valueOfModel := reflect.Indirect(reflect.ValueOf(model))
	modelType := valueOfModel.Type()
	idFieldName := FindIdField(modelType)
	idValue := valueOfModel.FieldByName(idFieldName).Interface()
	idType := valueOfModel.FieldByName(idFieldName).Type()
	model1 := reflect.New(modelType).Interface()
	DefaultObjID, _ := primitive.ObjectIDFromHex("000000000000")

	if idValue == "" || idValue == DefaultObjID {
		result, bool, err := InsertOne(ctx, collection, model)
		if bool {
			return nil, err
		} else {
			return result, nil
		}
	} else {
		if idType.String() == "primitive.ObjectID" {
			filter := bson.M{"_id": idValue}
			update := bson.M{
				"$set": model,
			}
			result := collection.FindOneAndUpdate(ctx, filter, update)
			err := result.Decode(&model1)
			return model1, err
		} else {
			isExisted, _ := Exist(ctx, collection, idValue, false)
			if isExisted {
				filter := bson.M{"_id": idValue}
				update := bson.M{
					"$set": model,
				}
				result := collection.FindOneAndUpdate(ctx, filter, update)
				err := result.Decode(&model1)
				return model1, err
			} else {
				result, bool, err := InsertOne(ctx, collection, model)
				if bool {
					return nil, err
				} else {
					return result, nil
				}
			}
		}
	}
}

func UpsertMany(ctx context.Context, collection *mongo.Collection, model interface{}) (interface{}, error) { //Patch
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
func FindIdField(modelType reflect.Type) string {
	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)
		bsonTag := field.Tag.Get("bson")
		tags := strings.Split(bsonTag, ",")
		for _, tag := range tags {
			if strings.Compare(strings.TrimSpace(tag), "_id") == 0 {
				return field.Name
			}
		}
	}
	return ""
}

//For Search and Patch
func GetBsonColumnName(ModelType reflect.Type, fieldName string) string {
	fields, _ := ModelType.FieldByName(fieldName)
	if tag, ok := fields.Tag.Lookup("bson"); ok {
		return tag
	} else {
		return fieldName
	}
}

//For Update
func BuildQueryByIdFromObject(object interface{}, idName string) bson.M {
	var query bson.M
	if v, ok := object.(map[string]interface{}); ok {
		query = bson.M{"_id": v[idName]}
	} else {
		value := reflect.Indirect(reflect.ValueOf(object)).FieldByName(idName).Interface()
		query = bson.M{"_id": value}
	}
	return query
}

//For Patch
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
	elemValue := arrValue.Elem()

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
			model, id := mapToMongoObject(values.Index(i).Interface(), idName, idObjectId, newId)
			ids = append(ids, id)
			results = appendToArray(results, model)
		}
	}
	return results, ids
}

func mapToMongoObject(model interface{}, idName string, IdObjectId bool, newId bool) (interface{}, interface{}) {
	id, _ := getValue(model, idName)
	if IdObjectId {
		if newId && (id == nil) {
			setValue(model, idName, bsonx.ObjectID(primitive.NewObjectID()))
		} else {
			objectId, err := primitive.ObjectIDFromHex(id.(string))
			if err == nil {
				setValue(model, idName, objectId)
			}
		}
	} else {
		setValue(model, idName, id)
	}
	return model, id
}

func getValue(model interface{}, fieldName string) (interface{}, error) {
	valueObject := reflect.Indirect(reflect.ValueOf(model))
	numField := valueObject.NumField()
	for i := 0; i < numField; i++ {
		if fieldName == valueObject.Type().Field(i).Name {
			return reflect.Indirect(valueObject).FieldByName(fieldName).Interface(), nil
		}
	}
	return nil, fmt.Errorf("Error no found field: " + fieldName)
}

func setValue(model interface{}, fieldName string, value interface{}) (interface{}, error) {
	valueObject := reflect.Indirect(reflect.ValueOf(model))
	numField := valueObject.NumField()
	for i := 0; i < numField; i++ {
		if fieldName == valueObject.Type().Field(i).Name {
			reflect.Indirect(valueObject).FieldByName(fieldName).Set(reflect.ValueOf(value))
			return model, nil
		}
	}
	return nil, fmt.Errorf("Error no found field: " + fieldName)
}
