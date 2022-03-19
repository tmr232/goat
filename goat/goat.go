package goat

import (
	"reflect"

	"github.com/urfave/cli"
)

type Optional[T any] struct {
	HasValue bool
	Value    T
}

type baseTypeGetter interface {
	baseType() reflect.Type
}

func (o Optional[T]) baseType() reflect.Type {
	return reflect.TypeOf(o.Value)
}

func Empty[T any]() Optional[T] {
	return Optional[T]{}
}

func Value[T any](value T) Optional[T] {
	return Optional[T]{Value: value, HasValue: true}
}

func makeFlag(field reflect.StructField) cli.Flag {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("name")
	if hasAlias {
		name = alias
	}
	usage := field.Tag.Get("usage")

	required := true

	fieldType := field.Type

	typeGetter, isOptional := reflect.New(fieldType).Elem().Interface().(baseTypeGetter)
	if isOptional {
		fieldType = typeGetter.baseType()
		required = false
	}

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

func buildFlags(argsType reflect.Type) (flags []cli.Flag) {
	if argsType.Kind() != reflect.Struct {
		panic("Args must be a struct type!")
	}

	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		_, embed := field.Tag.Lookup("embed")
		if embed {
			flags = append(flags, buildFlags(field.Type)...)
		} else {
			flags = append(flags, makeFlag(field))
		}
	}
	return
}

func wrapValue[T any](value T, optional bool) any {
	if optional {
		return Value(value)
	}
	return value
}

func getField(c *cli.Context, field reflect.StructField) (any, bool) {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("goat")
	if hasAlias {
		name = alias
	}

	fieldType := field.Type

	required := false

	typeGetter, isOptional := reflect.New(fieldType).Elem().Interface().(baseTypeGetter)
	if isOptional {
		fieldType = typeGetter.baseType()
		required = false
	}

	var getFlag func(name string) any

	switch fieldType.Kind() {
	case reflect.Bool:
		getFlag = func(name string) any { return wrapValue(c.Bool(name), isOptional) }
	case reflect.String:
		getFlag = func(name string) any { return wrapValue(c.String(name), isOptional) }
	case reflect.Int:
		getFlag = func(name string) any { return wrapValue(c.Int(name), isOptional) }
	case reflect.Int64:
		getFlag = func(name string) any { return wrapValue(c.Int64(name), isOptional) }
	case reflect.Uint:
		getFlag = func(name string) any { return wrapValue(c.Uint(name), isOptional) }
	case reflect.Uint64:
		getFlag = func(name string) any { return wrapValue(c.Uint64(name), isOptional) }
	case reflect.Float64:
		getFlag = func(name string) any { return wrapValue(c.Float64(name), isOptional) }
	case reflect.Slice:
		switch fieldType.Elem().Kind() {
		case reflect.String:
			getFlag = func(name string) any { return wrapValue(c.StringSlice(name), isOptional) }
		case reflect.Int:
			getFlag = func(name string) any { return wrapValue(c.IntSlice(name), isOptional) }
		case reflect.Int64:
			getFlag = func(name string) any { return wrapValue(c.Int64Slice(name), isOptional) }
		}
	}

	if required || c.IsSet(name) {
		return getFlag(name), true
	}

	return nil, false
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

func Command[Args any](name string, action func(Args) error, parts ...AppPart) CommandWrapper {
	cmd := CommandWrapper(cli.Command{
		Name:   name,
		Flags:  buildFlags(reflect.TypeOf(*new(Args))),
		Action: buildAction(action),
	})

	for _, p := range parts {
		switch p := p.(type) {
		case UsageWrapper:
			cmd.Usage = string(p)
		}
	}
	return cmd
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
