package filter

import (
	"errors"
	"reflect"
	"sync"
)

// Filter error collection
var (
	ErrSourceNotArray = errors.New("Source value is not an array")
	ErrFilterFuncNil  = errors.New("Filter function cannot be nil")
	ErrFilterNotFunc  = errors.New("Filter argument must be a function")
)

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
