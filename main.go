package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tmr232/goat/goat"
)

type HelloArgs struct {
	Name string `goat:"name"`
}

func Hello(args HelloArgs) error {
	fmt.Println("Hello ", args.Name)
	return nil
}

type GoodbyeArgs struct {
	Name string `goat:"name"`
}

func Goodbye(args GoodbyeArgs) error {
	fmt.Println("Goodbye ", args.Name)
	return nil
}

func main() {
	app := goat.App(
		"hello-world",
		goat.Command("hello", Hello),
		goat.Command("goodbye", Goodbye),
		goat.Group("say",
			goat.Command("hello", Hello),
			goat.Command("goodbye", Goodbye),
		),
	)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
