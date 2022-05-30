package main

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goat

func app(name string, goodbye bool) {
	if goodbye {
		fmt.Printf("Goodbye, %s.\n", name)
	} else {
		fmt.Printf("Hello, %s!\n", name)

	}
}

func main() {
	goat.Run(app)
}
