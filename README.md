# memdb

memdb is an easy-to-use in-memory database for Go programming language. It provides a simple and efficient way to store and retrieve data in memory, with features such as type safety, transactions, rich indexing, and sorting.

## Features

* **Type Safety** - memdb uses generics to provide strong type safety, reducing issues that can be found at compile time. This ensures that the data you are working with is of the correct type, and minimizes the chance of runtime errors.
* **Transactions** - Transactions in memdb can span multiple tables, and are applied atomically. This means that all changes made within a transaction are either applied together or not at all, ensuring data consistency and integrity.
* **Rich Indexing** - memdb supports single and compound indexes, allowing you to filter and sort data based on multiple properties. This makes it easy to retrieve specific data from the database quickly and efficiently.
* **Sorting** - memdb allows you to sort query results by different entry properties, making it easy to organize and present data in a specific order.

## Project status

Currently, this library is not stable and **should not be used for production**. It is still under development and may contain bugs or unfinished features. However, it is open for contributions, bugs reporting and feature requests.

## Installation

To install memdb, you can use the go get command:

```sh
go get github.com/knobz-io/memdb
```

Then, you can import the package in your Go code:

```go
import "github.com/knobz-io/memdb"
```

## Usage

To start using memdb, you'll need to first create a table schema for your data models using the `memdb.Table[V]` struct. This schema will define the structure of your data, and specify any indexes that you want to create on it. Once the table schema is created, you can use it to initialize a new `*memdb.DB` instance.


### Creating table schema

In order to use memdb, you'll need to define your data models in Go. These will be the types that you'll store in the database. For example, you could have a `User` struct like this:

```go
type UserStatus int

const (
    Active Status = iota
    Suspended
    Banned
)

type User struct {
    ID       int
    Status   Status
    Email    string
    FullName string
}
```

Once you have your data models defined, you can create a table schema for them using the `memdb.Table[V]` struct. The schema will specify how the data is stored in the database, and will define any indexes that you want to create on it. For example, you could create a `UserTable` schema like this::

```go
type UserTable struct {
    memdb.Table[*User]
    // indexes
    status   *memdb.IntIndex[*User]
    email    *memdb.StringIndex[*User]
    fullName *memdb.StringIndex[*User]
}

// ID is a helper function for preparing 
// user primary key from int
func (UserTable) ID(id int) memdb.Key {
    return memdb.IntKey(id)
}

func makeUserTable() UserTable {
    table := memdb.NewTable(func(usr *User) memdb.Key {
        return memdb.IntKey(usr.ID)
    })
    // preparing indexes
    table, status := memdb.IndexInt(func(usr *User) {
        return usr.Status
    })
    table, email := memdb.IndexString(func(usr *User) {
        return usr.Email
    })
    table, fullName := memdb.IndexString(func(usr *User) {
        return usr.FullName
    })
    return UserTable{
        Table:         Table,
        status:        status,
        email:         email,
        fullName:      fullName,
    }
}
```

### Initializing database

Once you have created a table schema, you can use it to initialize a new `*memdb.DB` instance. The `Init` function takes a variable number of table schemas as arguments, allowing you to create multiple tables in a single database.

```go
users := makeUserTable()
db, err := memdb.Init(users)
```

### Inserting data

To insert data into the database, you'll need to start a write transaction using the `db.WriteTx()` method. Once you have a transaction, you can use the `Set` or `SetMulti` method on the table schema to insert one or multiple entries into the table.

```go
// start a write transaction
tx := db.WriteTx()

// insert one entry
users.Set(tx, &User{
    ID:       1,
    Status:   Active,
    Email:    "john.doe@example.com",
    FullName: "John Doe",
})

// or insert multiple entires
users.SetMulti(tx, []*Users{
    {
        ID:       1,
        Status:   Active,
        Email:    "john.doe@example.com",
        FullName: "John Doe",
    },
    {
        ID:       2,
        Status:   Active,
        Email:    "matt.smith@example.com",
        FullName: "Matt Smith",
    },
})

// commit changes
tx.Commit()
```

### Deleting data

To delete data from the database, you'll need to start a write transaction using the `db.WriteTx()` method. Once you have a transaction, you can use the `Del` or `DelMulti` method on the table schema to delete one or multiple entries from the table.

```go
// start a write transaction
tx := db.WriteTx()

// delete single entry where ID=1
err := users.Del(tx, users.ID(1))

// or delete multiple entries where ID=1, ID=2
err := users.DelMulti(tx, []users.Key{
    users.ID(1),
    users.ID(2),
})

// commit changes
tx.Commit()
```

### Retrieving single entry

To retrieve a specific entry from the table using its primary key, you'll need to start a read-only transaction using the `db.ReadTx()` method. Once you have a transaction, you can use the `Get` method on the table schema to retrieve the entry.

```go
// start read-only transaction
tx := db.ReadTx()
// retrieving entry with ID=1
usr, err := users.Get(tx, users.ID(1))
if err == memdb.ErrNotFound {
    // handle not found
} else if err != nil {
    panic(err)
}
```

### Retrieving multiple entries

The `Select` method on the table schema can be used to retrieve multiple entries from the table. The returned value is a query object that can be used to filter and sort the results.

```go
// start read-only transaction
tx := db.ReadTx()

// retrieve all entries
list, err := users.Select(tx).All()
```

This will retrieve all entries from the table, without any filtering or sorting.

You can also use the `Page` method to retrieve a specific page of entries from the table.

```go
// return limited number of entries
// from the speciffic point
limit := 10
offset := 5
list, err := users.Select(tx).Page(limit, offset)
```

This will retrieve a specific page of the entries with `limit` number of entries starting from `offset`.

For iterating over the table entries one by one, you can create a cursor and iterate over it.

```go
// create new cursor
c, err := users.Select(tx).Cursor()
if err != nil {
    panic(err)
}
// iterate over table entries
for usr, ok := c.First(); ok; usr. ok = c.Next() {
    fmt.Println("id:", usr.ID, "email:", usr.Email)
}
```
This will allow you to retrieve entries one by one and do something with them. This can be useful for large tables that you don't want to load into memory all at once.

Please note that, these examples are for retriving all the entries, if you want to filter the entries based on certain condition, you can use the `Where` method on the query object.

### Filtering

To retrieve a set of entries from the table based on certain conditions, you can use the `Select` method on the table schema to create a query, and then use various filter methods to specify the conditions.

```go
// start read-only transaction
tx := db.ReadTx()
// filter all entries where Status=Active
list, err := users.Select(tx).
    Where(
        users.status.Is(Active),
    ).
    All()
```

### Sorting

To retrieve a set of entries from the table and sort them based on certain properties, you can use the `Select` method on the table schema to create a query, and then use various sort methods to specify the sorting order.

```go
// start read-only transaction
tx := db.ReadTx()
// return a list of all users sorted all users
// sorted by FullName property in ascending order
list, err := users.Select(tx).
    OrderBy(users.fullName.Asc()).
    All()
```

In this example, the `OrderBy` method is used to sort the query results by the `FullName` property in ascending order. The `Asc()` method is used to specify the sorting order.

You can chain multiple sorting conditions by adding them to `OrderBy` method.

```go
list, err := users.Select(tx).
    OrderBy(
        users.fullName.Asc(),
        users.email.Desc(),
    ).
    All()
```

This will sort the query results by the `FullName` property in ascending order, then if two or more entries have the same `FullName` it will sort them by the `Email` property in descending order.

### Pagination

To retrieve a specific page of entries from the table based on certain conditions, you can use the `Page` method on the table schema to create a query, and then use the `limit` and `offset` as arguments to specify the pagination.

```go
// page details
limit := 10
offset := 0

// start read-only transaction
tx := db.ReadTx()
list, err := users.Select(tx).
    Where(
        users.status.Is(Active),
    ).
    OrderBy(users.fullName.Asc()).
    Page(limit, offset) 
```



For more information on how to use memdb, please refer to the [Godoc](https://pkg.go.dev/github.com/knobz-io/memdb).

## Contributing

If you're interested in contributing to memdb, please read the [contributing guidelines](CONTRIBUTING.md) before submitting a pull request.

## License

memdb is licensed under the [MIT License](LICENSE).


