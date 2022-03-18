package goat

import (
	"github.com/urfave/cli"
	"reflect"
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

func MakeApp[Args any](app func(Args) error) *cli.App {
	appType := reflect.TypeOf(app)
	if appType.Kind() != reflect.Func {
		panic("Must be a function type!")
	}
	if appType.NumIn() != 1 {
		panic("Must take an arguments struct")
	}

	argsType := appType.In(0)
	return &cli.App{
		Flags:  buildFlags(argsType),
		Action: buildAction(app),
	}
}
