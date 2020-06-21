package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"reflect"
)

type DefaultViewService struct {
	Database   *mongo.Database
	Collection *mongo.Collection
	modelType  reflect.Type
	idName     string
	idObjectId bool
	idNames    []string
}

func NewDefaultViewService(db *mongo.Database, modelType reflect.Type, collectionName string, idObjectId bool) *DefaultViewService {
	idName := FindIdField(modelType)
	if len(idName) == 0 {
		log.Println(modelType.Name() + " repository can't use functions that need Id value (Ex GetById, ExistsById, Save, Update) because don't have any fields of " + modelType.Name() + " struct define _id bson tag.")
	}
	var idNames []string
	idNames = append(idNames, idName)
	return &DefaultViewService{db, db.Collection(collectionName), modelType, idName, idObjectId, idNames}
}

func (m *DefaultViewService) GetIdNames() []string {
	return m.idNames
}

func (m *DefaultViewService) GetAll(ctx context.Context) (interface{}, error) {
	return Find(ctx, m.Collection, bson.M{}, m.modelType)
}

func (m *DefaultViewService) GetById(ctx context.Context, id interface{}) (interface{}, error) {
	return FindOneWithId(ctx, m.Collection, id, m.idObjectId, m.modelType)
}

func (m *DefaultViewService) Exist(ctx context.Context, id interface{}) (bool, error) {
	return Exist(ctx, m.Collection, id, m.idObjectId)
}
