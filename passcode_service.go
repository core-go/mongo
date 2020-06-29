package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

type PasscodeService struct {
	collection *mongo.Collection
	passcodeName  string
	expiredAtName string
}

func NewPasscodeService(db *mongo.Database, collectionName string, passcodeName string, expiredAtName string) *PasscodeService {
	return &PasscodeService{db.Collection(collectionName), passcodeName, expiredAtName}
}

func NewDefaultPasscodeService(db *mongo.Database, collectionName string) *PasscodeService {
	return &PasscodeService{db.Collection(collectionName), "passcode", "expiredAt"}
}

func (s *PasscodeService) Save(ctx context.Context, id string, passcode string, expiredAt time.Time) (int64, error) {
	pass := make(map[string]interface{})
	pass["_id"] = id
	pass[s.passcodeName] = passcode
	pass[s.expiredAtName] = expiredAt
	idQuery := bson.M{"_id": id}
	return UpsertOne(ctx, s.collection, idQuery, pass)
}

func (s *PasscodeService) Load(ctx context.Context, id string) (string, time.Time, error) {
	idQuery := bson.M{"_id": id}
	x := s.collection.FindOne(ctx, idQuery)
	er1 := x.Err()
	if er1 != nil {
		if strings.Compare(fmt.Sprint(er1), "mongo: no documents in result") == 0 {
			return "", time.Now().Add(-24 * time.Hour), nil
		}
		return "", time.Now().Add(-24 * time.Hour), er1
	}
	k, er3 := x.DecodeBytes()
	if er3 != nil {
		return "", time.Now().Add(-24 * time.Hour), er3
	}

	code := strings.Trim(k.Lookup(s.passcodeName).String(), "\"")
	expiredAt := k.Lookup(s.expiredAtName).Time()
	return code, expiredAt, nil
}

func (s *PasscodeService) Delete(ctx context.Context, id string) (int64, error) {
	idQuery := bson.M{"_id": id}
	return DeleteOne(ctx, s.collection, idQuery)
}
