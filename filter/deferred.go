package filter

import (
	"reflect"
	"sync"
)

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
