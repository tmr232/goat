package goat

import (
	"reflect"

	"github.com/urfave/cli"
)

type Optional[T any] struct {
	HasValue bool
	Value    T
}

func Empty[T any]() Optional[T] {
	return Optional[T]{}
}

func Value[T any](value T) Optional[T] {
	return Optional[T]{Value: value, HasValue: true}
}

func makeFlag(field reflect.StructField) cli.Flag {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("goat")
	if hasAlias {
		name = alias
	}
	switch field.Type.Kind() {
	case reflect.Bool:
		return &cli.BoolFlag{Name: name, Required: true}
	case reflect.String:
		return &cli.StringFlag{Name: name, Required: true}
	}

	switch field.Type {
	case reflect.TypeOf(Optional[bool]{}):
		return &cli.BoolFlag{Name: name}
	case reflect.TypeOf(Optional[string]{}):
		return &cli.StringFlag{Name: name}
	}

	panic("Unsupported type!")
}

func buildFlags(argsType reflect.Type) (flags []cli.Flag) {
	if argsType.Kind() != reflect.Struct {
		panic("Args must be a struct type!")
	}

	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		flags = append(flags, makeFlag(field))
	}
	return
}

func getField(c *cli.Context, field reflect.StructField) (any, bool) {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("goat")
	if hasAlias {
		name = alias
	}
	switch field.Type.Kind() {
	case reflect.Bool:
		return c.Bool(name), true
	case reflect.String:
		return c.String(name), true
	}

	switch field.Type {
	case reflect.TypeOf(Optional[bool]{}):
		if c.IsSet(name) {
			return Value(c.Bool(name)), true
		}
		return nil, false
	case reflect.TypeOf(Optional[string]{}):
		if c.IsSet(name) {
			return Value(c.String(name)), true
		}
		return nil, false
	}
	panic("Invalid field type")
}

func buildAction[Args any](action func(Args) error) func(c *cli.Context) error {
	actionType := reflect.TypeOf(action)
	if actionType.Kind() != reflect.Func {
		panic("Must be a function type!")
	}
	if actionType.NumIn() != 1 {
		panic("Must take an arguments struct")
	}

	argsType := actionType.In(0)

	actionFunc := func(c *cli.Context) error {
		var args Args

		argsValue := reflect.ValueOf(&args)

		for i := 0; i < argsType.NumField(); i++ {
			field := argsType.Field(i)
			value, isSet := getField(c, field)
			if isSet {
				fieldValue := argsValue.Elem().Field(i)
				fieldValue.Set(reflect.ValueOf(value))
			}
		}

		return action(args)
	}

	return actionFunc
}

type AppPart interface {
	appPart()
}

type CommandWrapper cli.Command

func (c CommandWrapper) appPart() {}

type ActionFunction func(c *cli.Context) error

func (a ActionFunction) appPart() {}

func Command[Args any](name string, action func(Args) error) CommandWrapper {
	return CommandWrapper(cli.Command{
		Name:   name,
		Flags:  buildFlags(reflect.TypeOf(*new(Args))),
		Action: buildAction(action),
	})
}

func Group(name string, parts ...AppPart) (cmd CommandWrapper) {
	cmd.Name = name
	for _, p := range parts {
		switch p := p.(type) {
		case CommandWrapper:
			cmd.Subcommands = append(cmd.Subcommands, cli.Command(p))
		case ActionFunction:
			panic("Groups can't contain action functions!")
		case UsageWrapper:
			cmd.Usage = string(p)
		}
	}
	return
}

func Action[Args any](action func(Args) error) ActionFunction {
	return buildAction(action)
}

type UsageWrapper string

func (u UsageWrapper) appPart() {}
func Usage(usage string) UsageWrapper {
	return UsageWrapper(usage)
}

func App(name string, parts ...AppPart) (app *cli.App) {
	app = &cli.App{}
	app.Name = name
	for _, p := range parts {
		switch p := p.(type) {
		case CommandWrapper:
			app.Commands = append(app.Commands, cli.Command(p))
		case ActionFunction:
			app.Action = p
		case UsageWrapper:
			app.Usage = string(p)
		}
	}
	return
}
