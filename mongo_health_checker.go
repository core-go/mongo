package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type MongoHealthChecker struct {
	db      *mongo.Database
	name    string
	timeout time.Duration
}

func NewMongoHealthChecker(db *mongo.Database, name string, timeout time.Duration) *MongoHealthChecker {
	return &MongoHealthChecker{db: db, name: name, timeout: timeout}
}
func NewHealthChecker(db *mongo.Database, options ...string) *MongoHealthChecker {
	var name string
	if len(options) >= 1 && len(options[0]) > 0 {
		name = options[0]
	} else {
		name = "mongo"
	}
	return NewMongoHealthChecker(db, name, 4 * time.Second)
}

func (s *MongoHealthChecker) Name() string {
	return s.name
}

func (s *MongoHealthChecker) Check(ctx context.Context) (map[string]interface{}, error) {
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

func (s *MongoHealthChecker) Build(ctx context.Context, data map[string]interface{}, err error) map[string]interface{} {
	if err == nil {
		return data
	}
	data["error"] = err.Error()
	return data
}
