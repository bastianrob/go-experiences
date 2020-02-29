# Implementing Map In Go

> GO IS NOT A FUNCTIONAL LANGUAGE!

Whatever, I'll still do it just because…

---

## So what's the benefit of `Map`

Imagine an array of object with `N` number of entries.
You want each entry in the array to be processed, but each process takes time for around 3 seconds.
You starts writing the code, `imperatively`

```go
arr := [...]interface{..., ..., ..., ...}
for i, entry := range arr {
    doSomething(i, entry)
}
```

Now `doSomething` is a heavy lifter. If it takes around 3s to process each entry.
`Imperatively`, total time (`t`) it takes to process `N` number of array is

`t = N * 3s`

By having a `Map` function we can try to process each entry separately, in a `go routine`

## Parallel Map in Go

Imagine a function called `Map`. It takes 2 parameter:

* `Array` of any kind of object. Let's call it `source`
* `Function` which takes an entry of the `source`, and returns a new value / type after processing the entry. Let's call it `transform`

And returns:

* Another `array` of object which is the result of `function` `transforming` `source's` entry

```go
// ParallelMap an array of something into another thing using go routine
// Example:
//  Map([]int{1,2,3}, func(num int) int { return num+1 })
//  Results: []int{2,3,4}
func ParallelMap(source interface{}, transform interface{}) (interface{}, error) {
    // TODO:
}
```

### TODO-1. Ensure source's type

```go
sourceV := reflect.ValueOf(source)
kindOf := sourceV.Kind()
if kindOf != reflect.Slice && kindOf != reflect.Array {
    return nil, errors.New("Source value is not an array")
}
```

### TODO-2 Ensure `transform` is not nil and is a function

```go
if transform == nil {
    return nil, errors.New("Transform function cannot be nil")
}

tv := reflect.ValueOf(transform)
if tv.Kind() != reflect.Func {
    return nil, errors.New("Transform argument must be a function")
}
```

### TODO-3 Preparing result container

Here we encounter a difficulty:

We want to make result container, but we don't know what the results' array entry type is!
The one who knows what the result type is the caller, so we change our `ParallelMap` function signature to:

```go
func ParallelMap(source interface{}, transform interface{}, T reflect.Type) (interface{}, error) {
```

Add some validation and proceed to create result container

```go
if T == nil {
    return nil, errors.New("Map result type cannot be nil")
}

// kinda equivalent to = make([]T, srcV.Len())
result := reflect.MakeSlice(reflect.SliceOf(T), srcV.Len(), srcV.Cap())
```

### TODO-4 Making the loop

```go

// create a waitgroup with length = source array length
// we'll reduce the counter each time an entry finished processing
wg := &sync.WaitGroup{}
wg.Add(srcV.Len())

// for each entry in source array
for i := 0; i < srcV.Len(); i++ {
    // one go routine for each entry
    go func(idx int, entry reflect.Value) {
        //Call the transformation and store the result value
        tfResults := tv.Call([]reflect.Value{entry})

        //Store the transformation result into array of result
        resultEntry := result.Index(idx)
        if len(tfResults) > 0 {
            resultEntry.Set(tfResults[0])
        } else {
            resultEntry.Set(reflect.Zero(T))
        }

        //this go routine is done
        wg.Done()
    }(i, srcV.Index(i))
}
```

### TODO-5 Wait and return

```go
wg.Wait()
return result.Interface(), nil
```

### Stitching it all together

```go
// Map Error Collection
var (
    ErrMapSourceNotArray   = errors.New("Source value is not an array")
    ErrMapTransformNil     = errors.New("Transform function cannot be nil")
    ErrMapTransformNotFunc = errors.New("Transform argument must be a function")
    ErrMapResultTypeNil    = errors.New("Map result type cannot be nil")
)

// ParallelMap an array of something into another thing using go routine
// Example:
//  Map([]int{1,2,3}, func(num int) int { return num+1 }, reflect.Type(1))
//  Results: []int{2,3,4}
func ParallelMap(source interface{}, transform interface{}, T reflect.Type) (interface{}, error) {
    srcV := reflect.ValueOf(source)
    kind := srcV.Kind()
    if kind != reflect.Slice && kind != reflect.Array {
        return nil, ErrMapSourceNotArray
    }

    if transform == nil {
        return nil, ErrMapTransformNil
    }

    tv := reflect.ValueOf(transform)
    if tv.Kind() != reflect.Func {
        return nil, ErrMapTransformNotFunc
    }

    if T == nil {
        return nil, ErrMapResultTypeNil
    }

    // kinda equivalent to = make([]T, srcv.Len())
    result := reflect.MakeSlice(reflect.SliceOf(T), srcV.Len(), srcV.Cap())

    // create a waitgroup with length = source array length
    // we'll reduce the counter each time an entry finished processing
    wg := &sync.WaitGroup{}
    wg.Add(srcV.Len())
    for i := 0; i < srcV.Len(); i++ {
        // one go routine for each entry
        go func(idx int, entry reflect.Value) {
            //Call the transformation and store the result value
            tfResults := tv.Call([]reflect.Value{entry})

            //Store the transformation result into array of result
            resultEntry := result.Index(idx)
            if len(tfResults) > 0 {
                resultEntry.Set(tfResults[0])
            } else {
                resultEntry.Set(reflect.Zero(T))
            }

            //this go routine is done
            wg.Done()
        }(i, srcV.Index(i))
    }

    wg.Wait()
    return result.Interface(), nil
}
```

### Ensuring ParallelMap behavior with unit test

```go
func TestParallelMap(t *testing.T) {
    type args struct {
        arr       interface{}
        transform interface{}
        t         reflect.Type
    }
    tests := []struct {
        name    string
        args    args
        want    interface{}
        wantErr bool
    }{
        {
            name:    "Argument is not an array",
            args:    args{arr: 1, transform: nil, t: nil},
            want:    nil,
            wantErr: true,
        },
        {
            name:    "Transform function is nil",
            args:    args{arr: []int{1, 2, 3}, transform: nil, t: nil},
            want:    nil,
            wantErr: true,
        },
        {
            name:    "Transform is not a function",
            args:    args{arr: []int{1, 2, 3}, transform: 1, t: nil},
            want:    nil,
            wantErr: true,
        },
        {
            name:    "T is not supplied",
            args:    args{arr: []int{1, 2, 3}, transform: 1, t: nil},
            want:    nil,
            wantErr: true,
        },
        {
            name: "Valid transform",
            args: args{arr: []int{1, 2, 3}, transform: func(num int) int {
                return num + 1
            }, t: reflect.TypeOf(1)},
            want:    []int{2, 3, 4},
            wantErr: false,
        },
        {
            name: "Valid transform",
            args: args{arr: []int{1, 2, 3}, transform: func(num int) string {
                return strconv.Itoa(num)
            }, t: reflect.TypeOf("")},
            want:    []string{"1", "2", "3"},
            wantErr: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParallelMap(tt.args.arr, tt.args.transform, tt.args.t)
            if (err != nil) != tt.wantErr {
                t.Errorf("Map() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Map() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

And test results shows:

```bash
Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/arrayutil -run ^(TestParallelMap)$

ok      github.com/bastianrob/arrayutil    0.005s
Success: Tests passed.
```

### How does it fare

Well, the code is actually still an imperative style wrapped in a `ParallelMap` function and utilize `go routine` to achieve concurrency. (Maybe we should name it `ConcurrentMap` instead?)

But how does it compare with the imaginary case at the start of this non-sensical `readme`? Let's test it

```go
func BenchmarkParallelMap(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }
    transf := func(num int) int {
        //fake, long running operations
        time.Sleep(20 * time.Millisecond)
        return num + 1
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        ParallelMap(source, transf, reflect.TypeOf(1))
    }
}

func BenchmarkImperative(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }
    transf := func(num int) int {
        //fake, long running operations
        time.Sleep(20 * time.Millisecond)
        return num + 1
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for _, num := range source {
            transf(num)
        }
    }
}
```

Benchmark result shows:

```bash
BenchmarkParallelMap            100      22908438 ns/op       13801 B/op         305 allocs/op
PASS
ok      github.com/bastianrob/arrayutil    2.321s
Success: Benchmarks passed.

BenchmarkImperative               1      2251692199 ns/op      1240 B/op           7 allocs/op
PASS
ok      github.com/bastianrob/arrayutil    2.258s
Success: Benchmarks passed.
```

We can finish 100 `ParallelMap` to a `source`, while standard imperative style only finish 1
