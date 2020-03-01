package filter

import (
	"reflect"
)

// Filter an array without go routine
func Filter(source, filter interface{}) (interface{}, error) {
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

	// for each entry in source
	for i := 0; i < srcV.Len(); i++ {
		entry := srcV.Index(i)
		// call filter function via reflection, and check the result
		valid := fv.
			Call([]reflect.Value{entry})[0].
			Interface().(bool)

		// if result is valid, send the entry into queue
		// else, send zero value into queue
		if valid {
			appendResult := reflect.Append(ptrToElementOfSliceT, entry)
			ptrToElementOfSliceT.Set(appendResult)
		}
	}

	return ptrToElementOfSliceT.Interface(), nil
}
