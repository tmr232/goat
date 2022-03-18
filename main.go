package main

import (
	"fmt"
	"github.com/tmr232/goat/goat"
	"log"
	"os"
)

type SomeArgs struct {
	Name          string                `goat:"name"`
	Flag          goat.Optional[bool]   `goat:"flag"`
	YetAnotherVar goat.Optional[string] `goat:"yet-another-var"`
}

func Action(args SomeArgs) error {
	fmt.Println("Name: ", args.Name)
	fmt.Println("Flag: ", args.Flag)
	fmt.Println("YetAnotherVar: ", args.YetAnotherVar)
	return nil
}

func main() {
	app := goat.MakeApp(Action)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
