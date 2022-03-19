package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tmr232/goat/goat"
)

type HelloArgs struct {
	Name goat.Optional[string] `goat:"name"`
}

func Hello(args HelloArgs) error {
	fmt.Println("Hello ", args.Name)
	return nil
}

type GoodbyeArgs struct {
	Name string `goat:"name" usage:"the name to say goodbye to"`
}

func Goodbye(args GoodbyeArgs) error {
	fmt.Println("Goodbye ", args.Name)
	return nil
}

func main() {
	app := goat.App(
		"hello-world",
		goat.Command("hello", Hello),
		goat.Command("goodbye", Goodbye, goat.Usage("Says goodbye.")),
		goat.Group("say",
			goat.Command("hello", Hello),
			goat.Command("goodbye", Goodbye),
			goat.Usage("Says many things."),
		),
		goat.Usage("Greets things & people."),
	)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
