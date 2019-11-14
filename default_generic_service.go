package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type DefaultGenericService struct {
	*DefaultViewService
	maps map[string]string
}

func NewDefaultGenericService(db *mongo.Database, modelType reflect.Type, collectionName string, idObjectId bool) *DefaultGenericService {
	defaultViewService := NewDefaultViewService(db, modelType, collectionName, idObjectId)
	return &DefaultGenericService{defaultViewService, MakeMapBson(modelType)}
}

func (m *DefaultGenericService) Insert(ctx context.Context, model interface{}) (interface{}, bool, error) {
	return InsertOne(ctx, m.Collection, model)
}

func (m *DefaultGenericService) Update(ctx context.Context, model interface{}) (interface{}, error) {
	query := BuildQueryByIdFromObject(model, m.idName)
	return UpdateOne(ctx, m.Collection, model, query)
}

func (m *DefaultGenericService) Patch(ctx context.Context, model map[string]interface{}) (interface{}, error) {
	query := BuildQueryByIdFromObject(model, m.idName)
	return PatchOne(ctx, m.Collection, MapToBson(model, m.maps), query)
}

func (m *DefaultGenericService) Save(ctx context.Context, model interface{}) (interface{}, error) {
	return UpsertOne(ctx, m.Collection, model)
}

func (m *DefaultGenericService) Delete(ctx context.Context, id interface{}) (int64, error) {
	query := bson.M{"_id": id}
	return DeleteOne(ctx, m.Collection, query)
}
