package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type MongoHealthService struct {
	db      *mongo.Database
	name    string
	timeout time.Duration
}

func NewMongoHealthService(db *mongo.Database, name string, timeout time.Duration) *MongoHealthService {
	return &MongoHealthService{db, name, timeout}
}

func NewDefaultMongoHealthService(db *mongo.Database) *MongoHealthService {
	return &MongoHealthService{db, "mongo", 5 * time.Second}
}

func (s *MongoHealthService) Name() string {
	return s.name
}

func (s *MongoHealthService) Check(ctx context.Context) (map[string]interface{}, error) {
	cancel := func() {}
	if s.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
	}
	defer cancel()

	res := make(map[string]interface{})
	info := make(map[string]interface{})
	checkerChan := make(chan error)
	go func() {
		checkerChan <- s.db.RunCommand(ctx, bsonx.Doc{{"ping", bsonx.Int32(1)}}).Decode(&info)
	}()
	select {
	case err := <-checkerChan:
		return res, err
	case <-ctx.Done():
		return res, fmt.Errorf("timeout")
	}
}

func (s *MongoHealthService) Build(ctx context.Context, data map[string]interface{}, err error) map[string]interface{} {
	if err == nil {
		return data
	}
	data["error"] = err.Error()
	return data
}
