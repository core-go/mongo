# Mongo
Mongo is a library to wrap [Mongo Driver](go.mongodb.org/mongo-driver/mongo) with these purposes:
#### Simplified Database Operations
- Simplify common database operations, such as CRUD (Create, Read, Update, Delete) operations, transactions, and batch processing
#### Reduced Boilerplate Code
- Reduce boilerplate code associated with database interactions, allowing developers to focus more on application logic rather than low-level database handling
  - In this [go-mongo-sample](https://github.com/source-code-template/go-mongo-sample), you can see we can reduce a lot of source code at [data access layer](https://github.com/source-code-template/go-mongo-sample/blob/main/internal/user/repository/adapter/adapter.go), or you use [generic repository](https://github.com/core-go/mongo/blob/main/repository/repository.go) to replace all repository source code.
## Some advantage features
#### Generic CRUD Repository
[Repository](https://github.com/core-go/mongo/blob/main/repository/repository.go) is like [CrudRepository](https://docs.spring.io/spring-data/commons/docs/current/api/org/springframework/data/repository/CrudRepository.html) of Spring, it provides these advantages:
- <b>Simplicity</b>: provides a set of standard CRUD (Create, Read, Update, Delete) operations out of the box, reducing the amount of boilerplate code developers need to write.
    - Especially, it provides "Save" method, to build an insert or update statement, specified for Oracle, MySQL, MS SQL, Postgres, SQLite.
- <b>Consistency</b>: By using Repository, the code follows a consistent pattern for data access across the application, making it easier to understand and maintain.
- <b>Rapid Development</b>: reducing boilerplate code and ensuring transactional integrity.
- <b>Flexibility</b>: offers flexibility and control over complex queries, because it uses "database/sql" at GO SDK level.
- <b>Type Safety</b>: being a generic interface, it provides type-safe access to the entity objects, reducing the chances of runtime errors.
- <b>Learning Curve</b>: it supports utilities at GO SDK level. So, a developer who works with [Mongo Driver](go.mongodb.org/mongo-driver/mongo) can quickly understand and use it.
- <b>Conclusion</b>: The Repository offers a straightforward way to implement basic CRUD operations, promoting rapid development and consistency across applications. While it provides many advantages, such as reducing boilerplate code, it also it also offers flexibility and control over complex queries, because it uses [Mongo Driver](go.mongodb.org/mongo-driver/mongo) level.
- <b>Samples</b>: The sample is at [go-mongo-generic-sample](https://github.com/source-code-template/go-mongo-generic-sample).
#### Filtering, Pagination and Sorting
- <b>Filtering</b> is the process of narrowing down a dataset based on specific criteria or conditions. This allows users to refine the results to match their needs, making it easier to find relevant data.
- <b>Pagination</b> is the process of dividing a large dataset into smaller pages. Key Concepts of Pagination:
    - Page Size: The number of items displayed on each page.
        - Example: If you have 100 items and a page size of 10, there will be 10 pages in total.
    - Page Number: The current page being viewed.
        - Example: If you are on page 3 with a page size of 10, items 21 to 30 will be displayed.
    - Offset and Limit:
        - Offset: The number of items to skip before starting to collect the result set.
        - Limit: The maximum number of items to return.
        - Example: For page 3 with a page size of 10, the offset would be 20, and the limit would be 10 (SELECT * FROM items LIMIT 10 OFFSET 20).
- <b>Sorting</b>: build a dynamic SQL with sorting:
    - Build multi-column sorting based on dynamic parameters:
        - Input: sort=phone,-id,username,-dateOfBirth
        - Output: order by phone, id desc, username, date_of_birth desc
        - You can define your own format, and inject your own function to map
    - Safe and Secure Input Handling
        - See the above output, you can see we map JSON field name to database column name: username with username, dateOfBirth with date_of_birth
        - If you pass the columns which does not exist, the library ignore these columns.
#### Search Repository
Provide built-in support for pagination and filtering, making it easy to handle large datasets without writing additional code. The flow for search/paging:
- Build the dynamic filter
- Query data and map to array of struct
- Count the total of records for paging
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
- Sample is at [go-mongo-export](https://github.com/project-samples/go-mongo-export)
#### Mongo Health Check
- Sample is at [go-mongo-sample](https://github.com/source-code-template/go-mongo-sample)
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
