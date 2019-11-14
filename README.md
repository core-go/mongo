# Mongo
- Mongo Client Utilities

## Installation

Please make sure to initialize a Go module before installing common-go/mongo:

```shell
go get -u github.com/common-go/mongo
```

Import:

```go
import "github.com/common-go/mongo"
```

You can optimize the import by version:
- v1.0.0: Utilities to support query, find one by Id
- v1.0.1: Utilities to support insert, update, patch, upsert, delete

## Example

```go
type User struct {
	UserId    string `json:"id,omitempty" bson:"_id,omitempty"`
	UserName  string `json:"userName,omitempty" bson:"userName,omitempty"`
	Email     string `json:"email,omitempty" bson:"email,omitempty"`
	FirstName string `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty" bson:"lastName,omitempty"`
}

func main() {
	ctx := context.Background()
	db, _ := mongo.CreateConnection(ctx, "mongodb://localhost:27017", "master_data")
	collection := db.Collection("user")
	result := collection.FindOne(ctx, bson.M{"_id": "1484e7bd3e884971b3affa813bf30af0"})
	var user model.User
	result.Decode(&user)
	fmt.Println("email ", user.Email)
}
```
