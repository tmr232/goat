package goat

import (
	"reflect"

	"github.com/urfave/cli"
)

func makeFlag(field reflect.StructField) cli.Flag {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("name")
	if hasAlias {
		name = alias
	}
	usage := field.Tag.Get("usage")

	required := true

	fieldType := field.Type

	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
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

func shouldEmbed(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.Struct
}

func buildFlags(argsType reflect.Type) (flags []cli.Flag) {
	if argsType.Kind() != reflect.Struct {
		panic("Args must be a struct type!")
	}

	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		if shouldEmbed(field.Type) {
			flags = append(flags, buildFlags(field.Type)...)
		} else {
			flags = append(flags, makeFlag(field))
		}
	}
	return
}

func wrapValue[T any](value T, isPointer bool) any {
	if isPointer {
		return &value
	}
	return value
}

func getField(c *cli.Context, field reflect.StructField) (any, bool) {
	name := field.Name
	alias, hasAlias := field.Tag.Lookup("name")
	if hasAlias {
		name = alias
	}

	fieldType := field.Type

	required := false

	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
		required = false
	}
	isPointer := !required
	var getFlag func(name string) any

	switch fieldType.Kind() {
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
		switch fieldType.Elem().Kind() {
		case reflect.String:
			getFlag = func(name string) any { return wrapValue(c.StringSlice(name), isPointer) }
		case reflect.Int:
			getFlag = func(name string) any { return wrapValue(c.IntSlice(name), isPointer) }
		case reflect.Int64:
			getFlag = func(name string) any { return wrapValue(c.Int64Slice(name), isPointer) }
		}
	}

	if required || c.IsSet(name) {
		return getFlag(name), true
	}

	return nil, false
}

func setArgs(argsValue reflect.Value, c *cli.Context) {
	argsType := argsValue.Type()

	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		if shouldEmbed(field.Type) {
			setArgs(argsValue.Field(i), c)
		} else {
			value, isSet := getField(c, field)
			if isSet {
				fieldValue := argsValue.Field(i)
				fieldValue.Set(reflect.ValueOf(value))
			}
		}
	}
}

func buildAction[Args any](action func(Args) error) func(c *cli.Context) error {
	actionType := reflect.TypeOf(action)
	if actionType.Kind() != reflect.Func {
		panic("Must be a function type!")
	}
	if actionType.NumIn() != 1 {
		panic("Must take an arguments struct")
	}

	actionFunc := func(c *cli.Context) error {

		argsValue := reflect.New(actionType.In(0)).Elem()
		setArgs(argsValue, c)
		actionValue := reflect.ValueOf(action)
		ret := actionValue.Call([]reflect.Value{argsValue})[0].Interface()
		if ret != nil {
			return ret.(error)
		}
		return nil
	}

	return actionFunc
}

type CommandWrapper cli.Command

type ActionWrapper struct {
	action func(c *cli.Context) error
	flags  []cli.Flag
}

type UsageWrapper string

func Usage(usage string) UsageWrapper {
	return UsageWrapper(usage)
}

type AppPart interface {
	appPart()
}

func (c CommandWrapper) appPart() {}
func (a ActionWrapper) appPart()  {}
func (u UsageWrapper) appPart()   {}

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
		case ActionWrapper:
			panic("Groups can't contain action functions!")
		case UsageWrapper:
			cmd.Usage = string(p)
		}
	}
	return
}
func Action[Args any](action func(Args) error) ActionWrapper {
	return ActionWrapper{
		action: buildAction(action),
		flags:  buildFlags(reflect.TypeOf(*new(Args))),
	}
}

func App(name string, parts ...AppPart) (app *cli.App) {
	app = &cli.App{}
	app.Name = name
	for _, p := range parts {
		switch p := p.(type) {
		case CommandWrapper:
			app.Commands = append(app.Commands, cli.Command(p))
		case ActionWrapper:
			app.Action = p.action
			app.Flags = p.flags
		case UsageWrapper:
			app.Usage = string(p)
		}
	}
	return
}
