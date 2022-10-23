package tests

import "github.com/tmr232/goat"

//go:generate go run github.com/tmr232/goat/cmd/goater

func noFlags()         {}
func intFlag(flag int) {}
func renamedFlag(bla int) {
	goat.Flag(bla).Name("flag")
}

func Register() {
	goat.Command(noFlags)
	goat.Command(intFlag)
	goat.Command(renamedFlag)
}
