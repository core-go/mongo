package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type ServerInfoService struct {
	db      *mongo.Database
	name    string
	timeout time.Duration
}

func NewServerInfoService(db *mongo.Database, name string, timeout time.Duration) *ServerInfoService {
	return &ServerInfoService{db, name, timeout}
}

func NewDefaultServerInfoService(db *mongo.Database) *ServerInfoService {
	return &ServerInfoService{db, "mongo", 5 * time.Second}
}

func (s *ServerInfoService) Name() string {
	return s.name
}

func (s *ServerInfoService) Check(ctx context.Context) (map[string]interface{}, error) {
	cancel := func() {}
	if s.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
	}
	defer cancel()

	res := make(map[string]interface{})
	info := make(map[string]interface{})
	checkerChan := make(chan error)
	go func() {
		checkerChan <- s.db.RunCommand(ctx, bsonx.Doc{{"serverStatus", bsonx.Int32(1)}}).Decode(&info)
	}()
	select {
	case err := <-checkerChan:
		res["version"] = info["version"]
		return res, err
	case <-ctx.Done():
		return res, fmt.Errorf("timeout")
	}
}

func (s *ServerInfoService) Build(ctx context.Context, data map[string]interface{}, err error) map[string]interface{} {
	if err == nil {
		return data
	}
	data["error"] = err.Error()
	return data
}
