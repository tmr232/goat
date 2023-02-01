package tests

import (
	"fmt"
	"github.com/tmr232/goat"
	"io"
	"reflect"
)

//go:generate go run github.com/tmr232/goat/cmd/goater

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

func withIntFlags(required, defaultValue int, optional *int) {
	goat.Flag(defaultValue).Default(42)

	printArg("required", required)
	printArg("defaultValue", defaultValue)
	printArg("optional", optional)
}

func noFlags()         {}
func intFlag(flag int) {}
func renamedFlag(bla int) {
	goat.Flag(bla).Name("flag")
}

// Documented has some neat docs!
//
// It's just so nice to document your code.
func Documented() {}

func flagUsage(num int, str string) {
	goat.Flag(num).Usage("A number of things.")
	goat.Flag(str).Usage("A piece of text.")
}

func defaultValue(num int) {
	goat.Flag(num).Default(5)
}

func optionalFlag(num *int) {
	goat.Flag(num).Usage("This flag is optional!")

	if num == nil {
		fmt.Fprintln(testCommandWriter, "No value provided.")
	} else {
		fmt.Fprintln(testCommandWriter, *num)
	}
}

func Register() {
	goat.Command(noFlags)
	goat.Command(intFlag)
	goat.Command(renamedFlag)
	goat.Command(Documented)
	goat.Command(flagUsage)
	goat.Command(defaultValue)
	goat.Command(optionalFlag)
	goat.Command(withIntFlags)
}
