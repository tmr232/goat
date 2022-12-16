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

func optionalFlag(num *int) {
	goat.Flag(num).Usage("This flag is optional!")

	if num == nil {
		fmt.Printf("No value provided.")
	}
	fmt.Println(num)
}

func Register() {
	goat.Command(noFlags)
	goat.Command(intFlag)
	goat.Command(renamedFlag)
	goat.Command(Documented)
	goat.Command(flagUsage)
	goat.Command(defaultValue)
	goat.Command(optionalFlag)
}
