# Implementing Reduce In Go

> GO IS NOT A FUNCTIONAL LANGUAGE!

Whatever, I'll still do it just because…

---

## So what is `Reduce`

Imagine an array of `object` with `N` number of entries.
You want to calculate the `sum` and `average` value in the array and starts writing the code, `imperatively`/

```go
sum := 0
arr := [...]int{1, 3, 5, 7, 11}
for _, num := range arr {
    sum += num
}

avg := float32(sum) / float32(len(arr))
```

At a glance, its hard to know what the `fuck` is going on when using a loop to do data transformation.
`Sum` is calculated inside loop while `avg` calculated outside.
It can even be more complex when the data transformation is complicated and includes multiple branching condition.

> We often need to read all code inside a loop, just to know its intention.

## Reduce in Go

Imagine a function called `Reduce`. It takes 3 parameters:

* `Array` of any kind of object. Let's call it `source`
* `Initial Value` as the name implies, initial value of the expected result
* `Function` which takes 3 parameter and returns a value. Let's call it `reducer`

```go
//Reduce an array of something into another thing
func Reduce(source, initialValue, reducer interface{}) (interface{}, error) {
    //TODO:
}
```

Now you might be asking: why not use typed function for the `reducer`?

```go
// Reducer is a typed function which takes 3 parameter and must return a value
type Reducer func(accumulator, entry interface{}, idx int) interface{}
```

Now look at the code below

```go
notFun := Reducer(func(accumulator string, entry int, idx int) string {
    return fmt.Sprintln(accumulator, entry)
})
```

It will give you a compile error:

```text
cannot convert func literal (type func(string, int, int) string) to type Reducer
```

This stiff type checking is not what we want to have.
We want the `reducer` function to accept any type of `accumulator`, `entry` and returns whatever we need it to return.

### TODO-1. Ensure source's type

```go
srcV := reflect.ValueOf(source)
kind := srcV.Kind()
if kind != reflect.Slice && kind != reflect.Array {
    return nil, fmt.Errorf("Source value is not an array")
}
```

### TODO-2. Ensure `reducer` is not nil and is a function

```go
if reducer == nil {
    return nil, errors.New("Reducer function cannot be nil")
}

rv := reflect.ValueOf(reducer)
if rv.Kind() != reflect.Func {
    return nil, errors.New("Reducer argument must be a function")
}
```

### TODO-3 Preparing result container

```go
// copy initial value object
accumulator := initialValue
accV := reflect.ValueOf(initialValue)
```

### TODO-4 Making the loop and return the result

```go
// for each entry in source array
for i := 0; i < arrV.Len(); i++ {
    entry := arrV.Index(i)

    // call reducer via reflection,
    reduceResults := rv.Call([]reflect.Value{
        accV,               // send accumulator value
        entry,              // send current source entry
        reflect.ValueOf(i), // send current loop index
    })

    // get the first index of its result, and store it as accumulator value
    accV = reduceResults[0]
}

return accV.Interface(), nil
```

### Stitching it all together

```go
// Reducer Error Collection
var (
    ErrSourceNotArray   = errors.New("Source value is not an array")
    ErrReducerNil     = errors.New("Reducer function cannot be nil")
    ErrReducerNotFunc = errors.New("Reducer argument must be a function")
)

// Reduce an array of something into another thing
func Reduce(source, initialValue, reducer interface{}) (interface{}, error) {
    srcV := reflect.ValueOf(source)
    kind := srcV.Kind()
    if kind != reflect.Slice && kind != reflect.Array {
        return nil, ErrSourceNotArray
    }

    if reducer == nil {
        return nil, ErrReducerNil
    }

    rv := reflect.ValueOf(reducer)
    if rv.Kind() != reflect.Func {
        return nil, ErrReducerNotFunc
    }

    // copy initial value as accumulator, and get the reflection value
    accumulator := initialValue
    accV := reflect.ValueOf(accumulator)
    for i := 0; i < srcV.Len(); i++ {
        entry := srcV.Index(i)

        // call reducer via reflection
        reduceResults := rv.Call([]reflect.Value{
            accV,               // send accumulator value
            entry,              // send current source entry
            reflect.ValueOf(i), // send current loop index
        })

        accV = reduceResults[0]
    }

    return accV.Interface(), nil
}
```

### Ensuring Reduce behavior with unit test

```go
func TestReduce(t *testing.T) {
    type Person struct {
        Name       string
        Birthplace string
    }
    type PersonGroup map[string][]string
    type SumAvg struct {
        Sum int
        Avg float32
    }

    type args struct {
        source       interface{}
        initialValue interface{}
        reducer      interface{}
    }

    sumOfInt := func(accumulator, entry, idx int) int {
        return accumulator + entry
    }

    avgOfInt := func(accumulator SumAvg, entry, idx int) SumAvg {
        sum := accumulator.Sum + entry
        return SumAvg{
            Sum: sum,
            Avg: float32(sum) / float32(idx+1),
        }
    }

    groupBirthplacesByName := func(accumulator PersonGroup, entry Person, idx int) PersonGroup {
        birthplaces, exists := accumulator[entry.Name]
        if !exists {
            birthplaces = []string{entry.Birthplace}
        } else {
            birthplaces = append(birthplaces, entry.Birthplace)
        }
        accumulator[entry.Name] = birthplaces
        return accumulator
    }

    tests := []struct {
        name    string
        args    args
        want    interface{}
        wantErr bool
    }{
        {
            name:    "Source must be an array",
            args:    args{source: "something"},
            wantErr: true,
        },
        {
            name:    "Reducer must not be nil",
            args:    args{source: []int{1, 2, 3}, reducer: nil},
            wantErr: true,
        },
        {
            name:    "Reducer must be a function",
            args:    args{source: []int{1, 2, 3}, reducer: "something"},
            wantErr: true,
        },
        {
            name: "Sum of array",
            args: args{
                source:       []int{1, 2, 3},
                initialValue: 0,
                reducer:      sumOfInt,
            },
            wantErr: false,
            want:    6,
        },
        {
            name: "Avg of array",
            args: args{
                source:       []int{1, 2, 3},
                initialValue: SumAvg{Sum: 0, Avg: 0},
                reducer:      avgOfInt,
            },
            wantErr: false,
            want: SumAvg{
                Sum: 6,
                Avg: 6 / 3,
            },
        },
        {
            name: "Group by person's name",
            args: args{
                source: []Person{
                    Person{"John Doe", "Jakarta"},
                    Person{"John Doe", "Depok"},
                    Person{"John Doe", "Medan"},
                },
                initialValue: make(PersonGroup),
                reducer:      groupBirthplacesByName,
            },
            wantErr: false,
            want:    PersonGroup{"John Doe": []string{"Jakarta", "Depok", "Medan"}},
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Reduce(tt.args.source, tt.args.initialValue, tt.args.reducer)
            if (err != nil) != tt.wantErr {
                t.Errorf("Reduce() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Reduce() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

And test results shows:

```bash
Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/arrayutil -run ^(TestReduce)$

ok      github.com/bastianrob/arrayutil    0.005s
Success: Tests passed.
```

## Conclusion

By using reduce:

* We try to break down problem into smaller part
* So it can fit into one reducer function
* Which can be called `N` number of times
* To produce one result
* Instead of doing it all in a single imperative loop

It's also nice that we can:

* Read the reducer function name
* Get a quick glance of its intention
* Without having to read all the code inside
* And confident it is proucing correct result when the reducer function is properly unit tested
