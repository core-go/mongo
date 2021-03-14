# Mongo
#### Utilities
- Mongo Client Utilities
- HealthChecker and ServerInfoChecker
- FieldLoader
#### For Authentication, Sign in, Sign up, Password
- PasscodeRepository
#### For Batch Job
- Inserter
- Updater
- Upserter
- BatchInserter
- BatchUpdater
- BatchPatcher
- BatchWriter
#### For CRUD, search
- Loader
- Writer
- Searcher

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
- v0.0.2: HealthChecker and ServerInfoChecker
- v0.0.3: Utilities to support query, find one by Id, Mapper, FieldLoader, PasscodeRepository
- v0.0.4: Utilities to support insert, update, patch, upsert, delete
- v0.0.5: Utilities to support batch update
- v0.0.9: Loader, Writer, Inserter, Updater, Upserter, BatchInserter, BatchUpdater, BatchPatcher, BatchWriter
- v0.3.0: Searcher

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
