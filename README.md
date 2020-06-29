# Mongo
- Mongo Client Utilities
- HealthService
- FieldLoader
- ViewService
- GenericService
- SearchService

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
- v0.0.2: HealthService
- v0.0.3: Utilities to support query, find one by Id
- v0.0.5: Utilities to support insert, update, patch, upsert, delete
- v0.0.7: Utilities to support batch update
- v0.0.8: LocationMapper
- v0.0.9: FieldLoader 
- v1.0.6: LocationMapper, FieldLoader, ViewService and GenericService 
- v1.0.9: SearchService

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
