package cli

import (
	"reflect"

	"github.com/tmr232/goat/goat"
	"github.com/urfave/cli"
)

func MakeCliApp(goatApp goat.GoatApp) *cli.App {
	app := cli.App{
		Name:     goatApp.Name,
		Usage:    goatApp.Usage,
		Commands: makeCliCommands(goatApp.Commands),
	}
	if goatApp.Action != nil {
		app.Action = makeCliAction(goatApp.Action)
		app.Flags = makeCliFlags(goatApp.Action)
	}
	return &app
}

func makeCliCommands(goatCommands []goat.GoatCommand) (cliCommands cli.Commands) {
	for _, command := range goatCommands {
		switch command := command.(type) {
		case goat.GoatCommandSingle:
			cliCommands = append(cliCommands,
				cli.Command{
					Name:   command.Name,
					Usage:  command.Usage,
					Action: makeCliAction(&command.Action),
					Flags:  makeCliFlags(&command.Action),
				},
			)
		case goat.GoatCommandGroup:
			cliCommands = append(cliCommands,
				cli.Command{
					Name:        command.Name,
					Usage:       command.Usage,
					Subcommands: makeCliCommands(command.Subcommands),
				},
			)
		}
	}
	return
}

func makeCliFlags(action *goat.GoatAction) (flags []cli.Flag) {
	for _, goatFlag := range action.Flags {
		flags = append(flags, makeFlag(goatFlag))
	}
	return
}

func makeFlag(goatFlag goat.Flag) cli.Flag {
	name := goatFlag.DisplayName()
	usage := goatFlag.Usage
	required := goatFlag.Required
	fieldType := goatFlag.Type

	switch fieldType.Kind() {
	case reflect.Bool:
		return &cli.BoolFlag{Name: name, Required: required, Usage: usage}
	case reflect.String:
		return &cli.StringFlag{Name: name, Required: required, Usage: usage}
	case reflect.Int:
		return &cli.IntFlag{Name: name, Required: required, Usage: usage}
	case reflect.Int64:
		return &cli.Int64Flag{Name: name, Required: required, Usage: usage}
	case reflect.Uint:
		return &cli.UintFlag{Name: name, Required: required, Usage: usage}
	case reflect.Uint64:
		return &cli.Uint64Flag{Name: name, Required: required, Usage: usage}
	case reflect.Float64:
		return &cli.Float64Flag{Name: name, Required: required, Usage: usage}
	case reflect.Slice:
		switch fieldType.Elem().Kind() {
		case reflect.String:
			return &cli.StringSliceFlag{Name: name, Required: required, Usage: usage}
		case reflect.Int:
			return &cli.IntSliceFlag{Name: name, Required: required, Usage: usage}
		case reflect.Int64:
			return &cli.Int64SliceFlag{Name: name, Required: required, Usage: usage}
		}
	}

	panic("Unsupported type!")
}

type _Context struct {
	Context *cli.Context
}

func (context _Context) GetFlag(flag goat.Flag) (reflect.Value, bool) {
	// Make this work with the custom type, and then we're mostly set!
	c := context.Context
	name := flag.DisplayName()

	getFlag := func() reflect.Value {
		switch flag.Type.Kind() {
		case reflect.Bool:
			return reflect.ValueOf(c.Bool(name))
		case reflect.String:
			return reflect.ValueOf(c.String(name))
		case reflect.Int:
			return reflect.ValueOf(c.Int(name))
		case reflect.Int64:
			return reflect.ValueOf(c.Int64(name))
		case reflect.Uint:
			return reflect.ValueOf(c.Uint(name))
		case reflect.Uint64:
			return reflect.ValueOf(c.Uint64(name))
		case reflect.Float64:
			return reflect.ValueOf(c.Float64(name))
		case reflect.Slice:
			switch flag.Type.Elem().Kind() {
			case reflect.String:
				return reflect.ValueOf(c.StringSlice(name))
			case reflect.Int:
				return reflect.ValueOf(c.IntSlice(name))
			case reflect.Int64:
				return reflect.ValueOf(c.Int64Slice(name))
			}
		}
		panic("Why are we here?")
	}

	return getFlag(), c.IsSet(name)
}

func makeCliAction(goatAction *goat.GoatAction) cli.ActionFunc {
	actionFunc := func(c *cli.Context) error {
		return goatAction.Call(_Context{c})
	}

	return actionFunc
}
