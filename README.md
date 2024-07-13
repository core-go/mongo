# Mongo
- Mongo is a library to wrap [Mongo Driver](go.mongodb.org/mongo-driver/mongo) with these purposes:
#### Simplified Database Operations
- Simplify common database operations, such as CRUD (Create, Read, Update, Delete) operations, transactions, and batch processing
#### Reduced Boilerplate Code
- Reduce boilerplate code associated with database interactions, allowing developers to focus more on application logic rather than low-level database handling

## Some advantage features
#### Generic Repository (CRUD repository)
#### Search Repository
#### Dynamic query builder
#### For batch job
- Inserter
- Updater
- Writer
- StreamInserter
- StreamUpdater
- StreamWriter
- BatchInserter
- BatchUpdater
- BatchWriter
#### Geo Point Mapper
- Map latitude and longitude to mongo geo point
#### Export Service to export data
#### Firestore Health Check
#### Passcode Adapter
#### Activity Log
- Save Activity Log with dynamic database design
#### Field Loader

## Installation
Please make sure to initialize a Go module before installing core-go/mongo:

```shell
go get -u github.com/core-go/mongo
```

Import:
```go
import "github.com/core-go/mongo"
```
