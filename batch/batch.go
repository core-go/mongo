package batch

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertMany[T any](ctx context.Context, collection *mongo.Collection, objs []T) ([]int, error) {
	failIndices := make([]int, 0)
	arr := make([]interface{}, 0)
	l := len(objs)
	for i := 0; i < l; i++ {
		arr = append(arr, objs[i])
	}
	var defaultOrdered = false
	_, err := collection.InsertMany(ctx, arr, &options.InsertManyOptions{Ordered: &defaultOrdered})
	if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
		for _, writeError := range bulkWriteException.WriteErrors {
			failIndices = append(failIndices, writeError.Index)
		}
	}
	return failIndices, err
}
func UpdateMany[T any](ctx context.Context, collection *mongo.Collection, objs []T, opts ...int) (*mongo.BulkWriteResult, error) {
	le := len(objs)
	if le == 0 {
		return nil, nil
	}
	var index int
	if len(opts) > 0 && opts[0] >= 0 {
		index = opts[0]
	} else {
		var t T
		modelType := reflect.TypeOf(t)
		if modelType.Kind() != reflect.Struct {
			panic("T must be a struct")
		}
		index = FindIdField(modelType)
	}
	models := make([]mongo.WriteModel, 0)
	for i := 0; i < le; i++ {
		v := getValue(objs[i], index)
		updateQuery := bson.M{
			"$set": objs[i],
		}
		updateModel := mongo.NewUpdateOneModel().SetUpdate(updateQuery).SetFilter(bson.M{"_id": v})
		models = append(models, updateModel)
	}
	res, err := collection.BulkWrite(ctx, models)
	return res, err
}
func getValue(model interface{}, index int) interface{} {
	vo := reflect.ValueOf(model)
	return vo.Field(index).Interface()
}

// Patch
func PatchMaps(ctx context.Context, collection *mongo.Collection, maps []map[string]interface{}, idName string) (*mongo.BulkWriteResult, error) {
	if idName == "" {
		idName = "_id"
	}
	writeModels := make([]mongo.WriteModel, 0)
	for _, row := range maps {
		v, _ := row[idName]
		if v != nil {
			updateModel := mongo.NewUpdateOneModel().SetUpdate(bson.M{
				"$set": row,
			}).SetFilter(bson.M{"_id": v})
			writeModels = append(writeModels, updateModel)
		}
	}
	res, err := collection.BulkWrite(ctx, writeModels)
	return res, err
}
func UpsertMany[T any](ctx context.Context, collection *mongo.Collection, objs []T, opts ...int) (*mongo.BulkWriteResult, error) { //Patch
	le := len(objs)
	if le == 0 {
		return nil, nil
	}
	var index int
	if len(opts) > 0 && opts[0] >= 0 {
		index = opts[0]
	} else {
		var t T
		modelType := reflect.TypeOf(t)
		if modelType.Kind() != reflect.Struct {
			panic("T must be a struct")
		}
		index = FindIdField(modelType)
	}
	models := make([]mongo.WriteModel, 0)

	for i := 0; i < le; i++ {
		id := getValue(objs[i], index)
		if (reflect.TypeOf(id).String() == "string" && len(id.(string)) > 0) || !isNil(id) { // if exist
			updateModel := mongo.NewReplaceOneModel().SetUpsert(true).SetReplacement(objs[i]).SetFilter(bson.M{"_id": id})
			models = append(models, updateModel)
		} else {
			insertModel := mongo.NewInsertOneModel().SetDocument(objs[i])
			models = append(models, insertModel)
		}
	}
	res, err := collection.BulkWrite(ctx, models)
	return res, err
}
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
