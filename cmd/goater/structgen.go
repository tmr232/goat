package main

import (
	"reflect"
)

var flagByType map[string]reflect.Type

type baseFlag struct {
	Name string
}
type intFlag struct {
	baseFlag
}
type boolFlag struct {
	baseFlag
}

func addFlagHandler(typeName string, typ reflect.Type) {
	if flagByType == nil {
		flagByType = make(map[string]reflect.Type)
	}
	flagByType[typeName] = typ
}

func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf(*new(T))
}

func init() {
	addFlagHandler("int", TypeOf[intFlag]())
	addFlagHandler("bool", TypeOf[boolFlag]())
}

func getFlag(description FluentDescription) {
	/*
		1. Get the right type of flag struct
		2. Populate the relevant fields
		3. Return the struct!

		We can do it all with reflection to avoid really nasty codegen, and to allow custom types!
	*/
	rType, ok := flagByType[description.Type]
	if !ok {
		return
	}

	rVal := reflect.New(rType)
	rVal.FieldByName("Name").Set(reflect.ValueOf(description.Name))
	//for description.Descriptors
}
