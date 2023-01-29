package tests

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goater

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

func optionalFlag(num *int, ctx *goat.Context) {
	goat.Flag(num).Usage("This flag is optional!")

	if num == nil {
		fmt.Fprintf(ctx.GetWriter(), "No value provided.")
	} else {
		fmt.Fprintln(ctx.GetWriter(), *num)
	}
}

type customType int

func withCustomType(num customType) {}

func Register() {
	goat.Command(noFlags)
	goat.Command(intFlag)
	goat.Command(renamedFlag)
	goat.Command(Documented)
	goat.Command(flagUsage)
	goat.Command(defaultValue)
	goat.Command(optionalFlag)
	goat.Command(withCustomType)
}
