package cli

import (
	"github.com/tmr232/goat/goat"
	"github.com/urfave/cli"
	"reflect"
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

func makeCliAction(goatAction *goat.GoatAction) cli.ActionFunc {
	actionFunc := func(c *cli.Context) error {
		argsValue := reflect.New(goatAction.ArgsType).Elem()
		setArgs(argsValue, goatAction.Flags, c)
		actionValue := goatAction.ActionValue
		ret := actionValue.Call([]reflect.Value{argsValue})[0].Interface()
		if ret != nil {
			return ret.(error)
		}
		return nil
	}

	return actionFunc
}

func setArgs(argsValue reflect.Value, flags []goat.Flag, c *cli.Context) {
	for _, flag := range flags {
		value, isSet := getArg(c, flag)
		if isSet {
			argsValue.FieldByName(flag.ArgName()).Set(reflect.ValueOf(value))
		}
	}
}

func getArg(c *cli.Context, flag goat.Flag) (any, bool) {
	name := flag.Name
	isPointer := !flag.Required

	var getFlag func(name string) any

	switch flag.Type.Kind() {
	case reflect.Bool:
		getFlag = func(name string) any { return wrapValue(c.Bool(name), isPointer) }
	case reflect.String:
		getFlag = func(name string) any { return wrapValue(c.String(name), isPointer) }
	case reflect.Int:
		getFlag = func(name string) any { return wrapValue(c.Int(name), isPointer) }
	case reflect.Int64:
		getFlag = func(name string) any { return wrapValue(c.Int64(name), isPointer) }
	case reflect.Uint:
		getFlag = func(name string) any { return wrapValue(c.Uint(name), isPointer) }
	case reflect.Uint64:
		getFlag = func(name string) any { return wrapValue(c.Uint64(name), isPointer) }
	case reflect.Float64:
		getFlag = func(name string) any { return wrapValue(c.Float64(name), isPointer) }
	case reflect.Slice:
		switch flag.Type.Elem().Kind() {
		case reflect.String:
			getFlag = func(name string) any { return wrapValue(c.StringSlice(name), isPointer) }
		case reflect.Int:
			getFlag = func(name string) any { return wrapValue(c.IntSlice(name), isPointer) }
		case reflect.Int64:
			getFlag = func(name string) any { return wrapValue(c.Int64Slice(name), isPointer) }
		}
	}

	if flag.Required || c.IsSet(name) {
		return getFlag(name), true
	}

	return nil, false
}
func wrapValue[T any](value T, isPointer bool) any {
	if isPointer {
		return &value
	}
	return value
}
