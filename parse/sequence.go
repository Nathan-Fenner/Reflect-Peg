package parse

import "reflect"

// parseSequence is given a struct type representing a sequence.
func parseSequence(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	currentField := 0
	value := reflect.New(into).Elem()
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		fatal, ok := recovered.(fatalError)
		if !ok {
			panic(recovered) // keep unwinding without modification.
		}
		// We'll use Annotate now.
		for i := currentField - 1; i >= 0; i-- {
			method, ok := into.Field(i).Type.MethodByName("Annotate")
			if !ok {
				continue
			}
			if method.Type.NumIn() != 2 || method.Type.NumOut() != 1 {
				continue
			}
			if !reflect.TypeOf(fatal.Message).AssignableTo(method.Type.In(1)) {
				continue
			}
			output := method.Func.Call([]reflect.Value{value.Field(i), reflect.ValueOf(fatal.Message)})
			fatal.Message = output[0].Interface()
		}
		panic(fatal)
	}()
	for currentField < into.NumField() {
		result, err := parseIntoType(state, into.Field(currentField).Type, into.Field(currentField).Tag)
		if err != nil {
			return nil, err
		}
		value.Field(currentField).Set(reflect.ValueOf(result))
		currentField++
	}
	return value.Interface(), nil
}
