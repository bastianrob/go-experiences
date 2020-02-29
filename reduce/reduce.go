package reduce

import (
	"errors"
	"reflect"
)

// Reducer Error Collection
var (
	ErrSourceNotArray = errors.New("Source value is not an array")
	ErrReducerNil     = errors.New("Reducer function cannot be nil")
	ErrReducerNotFunc = errors.New("Reducer argument must be a function")
)

//Reduce an array of something into another thing
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
