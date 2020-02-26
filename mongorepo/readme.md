# MONGOREPO Abstraction Using `Generic Constructor`

Udah 1 tahun++ (sejak 2019) gw kerja pake golang + mongodb.
Library yang gw pake selama ini adalah:

<https://github.com/mongodb/mongo-go-driver>

## CRUD to resource

Most resource biasanya punya operasi CRUD yang kira-kira seperti:

```bash
'GET    /{resources}'       # Get a list of resource
'GET    /{resources}/{id}'  # Get one resource based on its ID
'POST   /{resource}'        # Create a new resource
'PUT    /{resource}/{id}'   # Update an existing resource based on its ID
'DELETE /{resource/{id}'    # Flag an existing resource as virtually deleted based on its ID
```

Which then brings us to next point

## Repository Design Pattern

Bayangin skenario dimana kita punya database object namanya `Person`

```go
type Person struct {
    ID   primitive.ObjecID `bson:"_id,omitempty"`
    Name string            `bson:"name"`
}
```

Dengan `Repository Design Pattern` biasanya kita akan define suatu repo misalnya:

```go
type PersonRepo struct {}

func(r *PersonRepo) Get(ctx context.Context) ([]*Person, error) {}
func(r *PersonRepo) GetOne(ctx context.Context, id string) (*Person, error) {}
func(r *PersonRepo) Create(ctx context.Context, p *Person) error {}
func(r *PersonRepo) Update(ctx context.Context, p *Person) error {}
func(r *PersonRepo) Delete(ctx context.Context, id string) error {}
```

dimana `PersonRepo` ini meng-handle 5 jenis `CRUD` yang kita define di point sebelumnya

## Abstraction, or the lack thereof

Bayangkan skenario dimana kita punya `buanyak (as in > 1)` database object selain `Person`, dan semuanya punya tipikal `CRUD` yang sama.
Di `C#` ato `Java` yang punya `generic` kita bisa bikin kira-kira kek gini:

```C#
public class Repo<T> {
    public T Get() {}
    public T GetOne(string id) {}
    public void Create(T obj) {}
    public void Update(T obj) {}
    public void Delete(string id) {}
}
```

Dimana `Repo<T>` bisa di inisialisasi di `runtime` untuk segala macam type/class kira-kira kek gini:

```C#
var personRepo = new Repo<Person>();
var enemyRepo = new Repo<Enemy>();
// etc, dst, dll, whatever
```

Tapi karena di `Golang`:

* Gak punya generic
* Bukan OOP

Kita (baca: software `engineer`/`developer`/`programmer`/`coder drones`/`highly trained monkeys`/`whatever`).
Kita terpaksa ngetik `type {Resource}Repo struct ...` 10 kali.

So, gimana caranya kita meng-abstract `Repo<T>` di `Golang`???

Pertama, kita define `struct` untuk `MongoRepo` beserta `factory` nya

```go
package mongorepo

// MongoRepo is repository that connects to MongoDB
// one instance of MongoRepo is responsible for one type of collection & data type
type MongoRepo struct {
    collection  *mongo.Collection
    constructor func() interface{}
}

// New creates a new instance of MongoRepo
func New(coll *mongo.Collection, cons func() interface{}) *MongoRepo {
    return &MongoRepo{
        collection: coll,
        constructor: cons,
    }
}
```

Di sini `MongoRepo` punya 2 field:

* `collection`: adalah mongodb collection mana yang akan di akses oleh repo
* `constructor`: adalah constructor/factory dari object yang mau kita abstraksi kan

So, think of the `constructor` as generic type `T` di `C#`.
Tapi, instead of storing the type information, we are storing the function on how to create a new object `T`

Untuk melihat gunanya `constructor` apaan, we move on to the `CRUD` implementation

```go
// Get a list of resource
// The function is simply getting all entries in r.collection for the sake of example simplicity
func (r *MongoRepo) Get(ctx context.Context) ([]interface{}, error) {
    cur, err := r.collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }

    var result []interface{}
    defer cur.Close(ctx)
    for cur.Next(ctx) {
        entry := r.constructor() // call to constructor
        if err = cur.Decode(entry); err != nil {
            return nil, err
        }

        result = append(result, entry)
    }

    return result, nil
}
```

Notice line 11 di snippet diatas
> `entry := r.constructor()`

Disini lah letak kita `ngakalin` abstraksi di `Golang`, yang gw sebut dengan `Generic Constructor`, gara-gara `Go` gak punya generic.

> `Efeknya apa sih? masih gak ngeh bray...`

Jadi, kita bisa bebas ngasih object/struct apapun di `constructor` function. e.g:

```go
// inisialisasi koneksi ke mongo
ctx := context.Background()
conn := os.Getenv("MONGO_CONN")
mongoopt := options.Client().ApplyURI(conn)
mongocl, _ := mongo.Connect(ctx, mongoopt)
mongodb := mongocl.Database("dbname")

// personRepo, ngarah ke collection 'person'
personRepo := mongorepo.New(
    mongodb.Collection("person"),
    func() interface{} {
        return &Person{}
    }))

// enemyRepo, ngarah ke collection 'enemy'
enemyRepo := mongorepo.New(
    mongodb.Collection("enemy"),
    func() interface{} {
        return &Enemy{}
    }))
```

Dengan contoh di atas, kita bandingin lagi dengan tandingannya di `C#` generic:

```C#
var personRepo = new Repo<Person>();
var enemyRepo = new Repo<Enemy>();
```

Udah cukup mirip lah ya gays.
Contoh code menyusul, kalo gw gak males. :v:
