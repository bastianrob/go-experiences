# Checking out Parallel Filter in Go

Imagine an array object with `N` number of entries.
You want to filter it to only get entries with `flagged = true`, OR `flagged = false, amount <= 500`.
You start writing the code, `imperatively`.

```go
type Object struct {
    Flagged bool
    Amount  int
}

var filtered []Object
arr := [...]Object{..., ..., ..., ...}
for i, entry := range arr {
    if !entry.Flagged && !(!entry.Flagged && entry.Amount <= 500) {
        continue
    }

    filtered = append(filtered, entry)
}
```

* Here, we need to read all code inside the loop, just to know its intention.
* Say if filtering each entry takes `x` amount of time, total time filtering takes with `n` number of entries is `t = n * x`
* Given `x=1s` and `n=1000` then `t=1000*1s = 1000s`

## Parallel Filtering in Go

Imagine a function called `ParallelFilter`. It takes 2 parameter:

* `Array` of any kind of object. Let's call it `source`
* `Function` which takes each entry of `Array` and returns a `boolean`. Let's call it `filter`

```go
// ParallelFilter an array using go routine
// This function will not guarantee order of results
func ParallelFilter(arr, filter interface{}) (interface{}, error) {
    //TODO:
}
```

### Ensuring ParallelFilter behavior with unit test

```go
func TestParallelFilter(t *testing.T) {
    intptr := func(num int) *int {
        return &num
    }
    type args struct {
        arr     interface{}
        filterf interface{}
    }
    tests := []struct {
        name    string
        args    args
        wantErr bool
        want    interface{}
    }{
        {"Success", args{
            arr: []int{1, 2, 3, 4},
            filterf: func(entry int) bool {
                return entry == 1
            }}, false, []int{1}},
        {"Success", args{
            arr: []*int{intptr(1), intptr(2), intptr(3), intptr(4)},
            filterf: func(entry *int) bool {
                return *entry == 1
            }}, false, []*int{intptr(1)}},
        {"Failed", args{
            arr:     "[]int{1, 2, 3, 4}",
            filterf: nil}, true, nil},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParallelFilter(tt.args.arr, tt.args.filterf)
            if (err != nil) != tt.wantErr {
                t.Errorf("Filter() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Filter() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### TODO-1. Ensure source's type

```go
srcV := reflect.ValueOf(source)
kind := srcV.Kind()
if kind != reflect.Slice && kind != reflect.Array {
    return nil, errors.New("Source value is not an array")
}
```

### TODO-2. Ensure `filter` is not nil and is a function

```go
if filter == nil {
    return nil, errors.New("Filter function cannot be nil")
}

fv := reflect.ValueOf(filter)
if fv.Kind() != reflect.Func {
    return nil, errors.New("Filter argument must be a function")
}
```

### TODO-3. Preparing result queue

Here's the basic idea, each time we found that entry satisfies `filter` function, we'll put the value into a `queue`.
There's then another process which waits for the `queue` and collect all the results.

```go
T := reflect.TypeOf(source).Elem()                      // 1. Get type T of source's element
sliceOfT := reflect.MakeSlice(reflect.SliceOf(T), 0, 0) // 2. var sliceOfT = new Slice<T>()
ptrToSliceOfT := reflect.New(sliceOfT.Type())           // 3. ptrToSliceOfT = &sliceOfT
ptrToElementOfSliceT := ptrToSliceOfT.Elem()            // 4. ptrToElementOfSliceT = &ptrToSliceOfT
// God damn it go! if only you know this thing called generic...

// Create a waitgroup with length = length of source's array
wg := &sync.WaitGroup{}
wg.Add(srcV.Len())

// make a buffered channel which collects valid filtered entries
queue := make(chan *reflect.Value, 3)

// This is a process that waits for queue and append it to result slice
go func() {
    // for each entry in the queue
    for entry := range queue {
        // if entry is not nil colect the entry and append it to the result container
        if entry != nil {
            appendResult := reflect.Append(ptrToSliceOfT.Elem(), *entry)
            ptrToSliceOfT.Elem().Set(appendResult)
        }
        // inform waitgroup that we finished 1 operation
        wg.Done()
    }
}()
```

### TODO-4 Making the loop and return the result

```go
// for each entry in source
for i := 0; i < srcV.Len(); i++ {
    // asynchronously check each entry
    go func(idx int, entry reflect.Value) {
        // call filter function via reflection, and check the result
        valid := fv.
            Call([]reflect.Value{entry})[0].
            Interface().(bool)

        // if result is valid, send the entry into queue
        // else, send zero value into queue
        if valid {
            queue <- &entry
        } else {
            queue <- nil
        }
    }(i, srcV.Index(i))
}

wg.Wait()    // wait for all filter to be done, and results appended to sliceValuePtr
close(queue) // close the queue channel so queue processor goroutine can exit
return ptrToElementOfSliceT.Interface(), nil
```

### Stitching it all together

```go
// ParallelFilter an array using go routine
// This function will not guarantee order of results
func ParallelFilter(source, filter interface{}) (interface{}, error) {
    srcV := reflect.ValueOf(source)
    kind := srcV.Kind()
    if kind != reflect.Slice && kind != reflect.Array {
        return nil, ErrSourceNotArray
    }

    if filter == nil {
        return nil, ErrFilterFuncNil
    }

    fv := reflect.ValueOf(filter)
    if fv.Kind() != reflect.Func {
        return nil, ErrFilterNotFunc
    }

    T := reflect.TypeOf(source).Elem()                      // 1. Get type T of source's element
    sliceOfT := reflect.MakeSlice(reflect.SliceOf(T), 0, 0) // 2. var sliceOfT = new Slice<T>()
    ptrToSliceOfT := reflect.New(sliceOfT.Type())           // 3. ptrToSliceOfT = &sliceOfT
    ptrToElementOfSliceT := ptrToSliceOfT.Elem()            // 4. ptrToElementOfSliceT = *ptrToSliceOfT
    // God damn it go! if only you know this thing called generic...

    // Create a waitgroup with length = length of source's array
    wg := &sync.WaitGroup{}
    wg.Add(srcV.Len())

    // make a buffered channel which collects valid filtered entries
    queue := make(chan *reflect.Value, 3)

    // This is a process that waits for queue and append it to result slice
    go func() {
        for entry := range queue {
            if entry != nil {
                appendResult := reflect.Append(ptrToElementOfSliceT, *entry)
                ptrToElementOfSliceT.Set(appendResult)
            }
            wg.Done()
        }
    }()

    // for each entry in source
    for i := 0; i < srcV.Len(); i++ {
        // asynchronously check each entry
        go func(idx int, entry reflect.Value) {
            // call filter function via reflection, and check the result
            valid := fv.
                Call([]reflect.Value{entry})[0].
                Interface().(bool)

            // if result is valid, send the entry into queue
            // else, send zero value into queue
            if valid {
                queue <- &entry
            } else {
                queue <- nil
            }
        }(i, srcV.Index(i))
    }

    wg.Wait()    // wait for all filter to be done, and results appended to sliceValuePtr
    close(queue) // close the queue channel so queue processor goroutine can exit
    return ptrToElementOfSliceT.Interface(), nil
}
```

And test results shows:

```bash
Running tool: /usr/local/opt/go/libexec/bin/go test -timeout 30s github.com/bastianrob/go-experiences/filter -run ^(TestParallelFilter)$

ok       github.com/bastianrob/go-experiences/filter    0.007s
Success: Tests passed.
```

### Benchmarking ParallelFilter

First we'll try to simulate slow running filter function

```go
func BenchmarkParallelFilter(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }
    isMultipliedBy3 := func(num int) bool {
        time.Sleep(20 * time.Millisecond) //simulate slow running filter
        return num%3 == 0
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        filter.ParallelFilter(source, isMultipliedBy3)
    }
}

func BenchmarkImperative(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }

    isMultipliedBy3 := func(num int) bool {
        time.Sleep(20 * time.Millisecond) //simulate slow running filter
        return num%3 == 0
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for _, num := range source {
            isMultipliedBy3(num)
        }
    }
}
```

Benchmark results:

```bash
BenchmarkParallelFilter-4            100      22920607 ns/op       18678 B/op         451 allocs/op
PASS
ok      github.com/bastianrob/go-experiences/filter    2.322s
Success: Benchmarks passed.

pkg: github.com/bastianrob/go-experiences/filter
BenchmarkImperative-4                  1    2289902017 ns/op        1240 B/op           7 allocs/op
PASS
ok      github.com/bastianrob/go-experiences/filter    2.295s
Success: Benchmarks passed.
```

`ParallelFilter` is faster when processing lots of entries with slow `filter` process.
But what about faster running `filter`?

```go
func BenchmarkParallelFilterFast(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }
    isMultipliedBy3 := func(num int) bool {
        return num%3 == 0
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        filter.ParallelFilter(source, isMultipliedBy3)
    }
}

func BenchmarkImperativeFast(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }

    isMultipliedBy3 := func(num int) bool {
        return num%3 == 0
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for _, num := range source {
            isMultipliedBy3(num)
        }
    }
}
```

And benchmark results:

```bash
BenchmarkParallelFilterFast-4        10000          184717 ns/op       11216 B/op         347 allocs/op
PASS
ok      github.com/bastianrob/go-experiences/filter    1.870s
Success: Benchmarks passed.

BenchmarkImperativeFast-4         50000000            40.7 ns/op           0 B/op           0 allocs/op
PASS
ok      github.com/bastianrob/go-experiences/filter    2.079s
Success: Benchmarks passed.
```

Shows that `ParallelFilter` is absolutely murdered by normal `imperative` filter.

## ParallelFilter slightly different approach

What if we inverse the responsibility of collecting filter result to the function caller?
This is possible in go by using channel so let's make `DeferredFilter` function:

```go
// DeferredFilter an array using go routine
// This function will not guarantee order of results
func DeferredFilter(source, filter interface{}) (<-chan interface{}, error) {
    srcV := reflect.ValueOf(source)
    kind := srcV.Kind()
    if kind != reflect.Slice && kind != reflect.Array {
        return nil, ErrSourceNotArray
    }

    if filter == nil {
        return nil, ErrFilterFuncNil
    }

    fv := reflect.ValueOf(filter)
    if fv.Kind() != reflect.Func {
        return nil, ErrFilterNotFunc
    }

    // Create a waitgroup with length = length of source's array
    wg := &sync.WaitGroup{}
    wg.Add(srcV.Len())

    // make a buffered channel which collects valid filtered entries
    queue := make(chan interface{}, 3)

    // for each entry in source
    for i := 0; i < srcV.Len(); i++ {
        // asynchronously check each entry
        go func(idx int, entry reflect.Value) {
            defer wg.Done()
            // call filter function via reflection, and check the result
            valid := fv.
                Call([]reflect.Value{entry})[0].
                Interface().(bool)

            // if result is valid, send the entry into queue
            // else, send zero value into queue
            if valid {
                queue <- &entry
            }
        }(i, srcV.Index(i))
    }

    go func() {
        wg.Wait()    // wait for all filter to be done, and results appended to sliceValuePtr
        close(queue) // close the queue channel so queue processor goroutine can exit
    }()

    return queue, nil
}
```

Benchmark it:

```go
func BenchmarkDeferredFilterFast(b *testing.B) {
    source := [100]int{}
    for i := 0; i < len(source); i++ {
        source[i] = i + 1
    }
    isMultipliedBy3 := func(num int) bool {
        return num%3 == 0
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        q, _ := filter.DeferredFilter(source, isMultipliedBy3)
        for _ = range q {
        }
    }
}
```

And the results is:

```bash
BenchmarkDeferredFilterFast-4          10000        144723 ns/op        9060 B/op         304 allocs/op
PASS
ok      github.com/bastianrob/go-experiences/filter    1.471s
Success: Benchmarks passed.
```

Nope... just slightly faster time / op

## Conclusion

Unless you are process big chunks of array and filtering each entry takes a long time.
Unfortunately there is no benefit to be gained by using `ParallelFilter`.
