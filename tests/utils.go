package tests

import (
	"fmt"
	"io"
	"reflect"
)

var testCommandWriter io.Writer

type writerContext struct {
	io.Writer
}

func (c writerContext) restore() {
	testCommandWriter = c
}

func withWriter(writer io.Writer) writerContext {
	original := testCommandWriter
	testCommandWriter = writer
	return writerContext{original}
}

func printArg(name string, value any) {
	if reflect.TypeOf(value).Kind() != reflect.Pointer || reflect.ValueOf(value).IsNil() {
		fmt.Fprintf(testCommandWriter, "%s = %#v\n", name, value)
	} else {
		fmt.Fprintf(testCommandWriter, "%s = *-> %#v\n", name, reflect.ValueOf(value).Elem().Interface())
	}
}
