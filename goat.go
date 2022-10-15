package goat

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"reflect"
)

type RunConfig struct {
	Flags []cli.Flag
	// TODO: Replace this with a function that takes the action function and returns ActionFunc
	//		 This is needed for supporting function literals instead of named functions.
	//		 This is also required for anything beyond named functions.
	Action         cli.ActionFunc
	CtxFlagBuilder func(c *cli.Context) map[string]any
	Name           string
	Usage          string
}

var runConfigByFunction map[reflect.Value]RunConfig
var functionByCliActionFunc map[reflect.Value]reflect.Value

func init() {
	runConfigByFunction = make(map[reflect.Value]RunConfig)
	functionByCliActionFunc = make(map[reflect.Value]reflect.Value)
}

// Register registers a RunConfig generated from a function.
//
// This is only used in generated code.
func Register(app any, config RunConfig) {
	appValue := reflect.ValueOf(app)
	runConfigByFunction[appValue] = config
	functionByCliActionFunc[reflect.ValueOf(config.Action)] = appValue
}

// RunE takes a free function and runs it as a CLI app.
func RunE(f any) error {
	config := runConfigByFunction[reflect.ValueOf(f)]

	app := &cli.App{
		Flags:  config.Flags,
		Action: config.Action,
		Name:   config.Name,
		Usage:  config.Usage,
	}

	return app.Run(os.Args)
}

// Run takes a free function and runs it as a CLI app, terminating with a log if an error occurs.
func Run(f any) {
	err := RunE(f)
	if err != nil {
		log.Fatal(err)
	}
}

func Command(f any, subcommands ...*cli.Command) *cli.Command {
	config := runConfigByFunction[reflect.ValueOf(f)]

	return &cli.Command{
		Flags:       config.Flags,
		Action:      config.Action,
		Name:        config.Name,
		Usage:       config.Usage,
		Subcommands: subcommands,
	}
}

type Application struct{ *cli.App }

func (app Application) RunE() error {
	return app.App.Run(os.Args)
}
func (app Application) Run() {
	err := app.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
func App(name string, commands ...*cli.Command) Application {
	return Application{
		App: &cli.App{
			Name:     name,
			Commands: commands,
		},
	}
}

// Flag creates a flag-descriptor to be used during code-generation to describe the flag.
//
// Should be used with the FluentFlag.Name, FluentFlag.Usage and FluentFlag.Default methods.
//
// Example:
// 	func f(myFlag int) {
//		Flag(myFlag).
//			Name("my-flag").
//			Usage("Just a flag.")
//	}
func Flag(any) FluentFlag {
	return FluentFlag{}
}

type FluentFlag struct{}

// Name sets the name of a flag.
func (f FluentFlag) Name(string) FluentFlag {
	return FluentFlag{}
}

// Usage sets the usage of a flag.
func (f FluentFlag) Usage(string) FluentFlag {
	return FluentFlag{}
}

// Default sets the default value for a flag.
//
// Must be called with the same type as the flag.
func (f FluentFlag) Default(any) FluentFlag {
	return FluentFlag{}
}

type FluentSelf struct{}

// Self begins a description-chain for the current function.
func Self() FluentSelf {
	return FluentSelf{}
}

// Name sets the name of the current function.
func (s FluentSelf) Name(string) FluentSelf {
	return FluentSelf{}
}

// Usage sets the usage of the current function.
func (s FluentSelf) Usage(string) FluentSelf {
	return FluentSelf{}
}
