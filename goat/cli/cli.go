package cli

import (
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

	switch goatFlag.Type {
	case goat.Bool:
		return &cli.BoolFlag{Name: name, Required: required, Usage: usage}
	case goat.String:
		return &cli.StringFlag{Name: name, Required: required, Usage: usage}
	case goat.Int:
		return &cli.IntFlag{Name: name, Required: required, Usage: usage}
	case goat.Int64:
		return &cli.Int64Flag{Name: name, Required: required, Usage: usage}
	case goat.Uint:
		return &cli.UintFlag{Name: name, Required: required, Usage: usage}
	case goat.Uint64:
		return &cli.Uint64Flag{Name: name, Required: required, Usage: usage}
	case goat.Float64:
		return &cli.Float64Flag{Name: name, Required: required, Usage: usage}
	case goat.StringSlice:
		return &cli.StringSliceFlag{Name: name, Required: required, Usage: usage}
	case goat.IntSlice:
		return &cli.IntSliceFlag{Name: name, Required: required, Usage: usage}
	case goat.Int64Slice:
		return &cli.Int64SliceFlag{Name: name, Required: required, Usage: usage}
	}

	panic("Unsupported type!")
}

type _Context struct {
	Context *cli.Context
}

func (context _Context) GetFlag(flag goat.Flag) (any, bool) {
	// Make this work with the custom type, and then we're mostly set!
	c := context.Context
	name := flag.DisplayName()

	getFlag := func() any {
		switch flag.Type {
		case goat.Bool:
			return c.Bool(name)
		case goat.String:
			return c.String(name)
		case goat.Int:
			return c.Int(name)
		case goat.Int64:
			return c.Int64(name)
		case goat.Uint:
			return c.Uint(name)
		case goat.Uint64:
			return c.Uint64(name)
		case goat.Float64:
			return c.Float64(name)
		case goat.StringSlice:
			return c.StringSlice(name)
		case goat.IntSlice:
			return c.IntSlice(name)
		case goat.Int64Slice:
			return c.Int64Slice(name)
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
