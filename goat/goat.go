package goat

import (
	"reflect"
)

type UsageWrapper string

func Usage(usage string) UsageWrapper {
	return UsageWrapper(usage)
}

type AppPart interface {
	appPart()
}

func (g GoatCommandSingle) appPart() {}
func (g GoatCommandGroup) appPart()  {}
func (a GoatAction) appPart()        {}
func (u UsageWrapper) appPart()      {}

func Command[Args any](name string, action func(Args) error, parts ...AppPart) GoatCommandSingle {
	cmd := GoatCommandSingle{
		Name:   name,
		Action: Action(action),
	}

	for _, p := range parts {
		switch p := p.(type) {
		case UsageWrapper:
			cmd.Usage = string(p)
		}
	}
	return cmd
}

func Group(name string, parts ...AppPart) GoatCommandGroup {
	cmd := GoatCommandGroup{
		Name: name,
	}
	for _, p := range parts {
		switch p := p.(type) {
		case GoatCommand:
			cmd.Subcommands = append(cmd.Subcommands, p)
		case UsageWrapper:
			cmd.Usage = string(p)
		}
	}
	return cmd
}

func Flags(argsType reflect.Type) (flags []Flag) {
	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)

		if shouldEmbed(field.Type) {
			flags = append(flags, Flags(field.Type)...)
		} else {
			name := field.Name
			alias := field.Tag.Get("alias")
			usage := field.Tag.Get("usage")

			required := true
			fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
				required = false
			}

			flag := Flag{
				Name:     name,
				Alias:    alias,
				Usage:    usage,
				Type:     fieldType,
				Required: required,
			}
			flags = append(flags, flag)
		}
	}
	return
}

func Action[Args any](action func(Args) error) GoatAction {
	return GoatAction{
		ActionValue: reflect.ValueOf(action),
		ArgsType:    reflect.TypeOf(action).In(0),
		Flags:       Flags(reflect.TypeOf(action).In(0)),
	}
}

func App(name string, parts ...AppPart) GoatApp {
	app := GoatApp{
		Name: name,
	}
	for _, p := range parts {
		switch p := p.(type) {
		case GoatCommand:
			app.Commands = append(app.Commands, p)
		case GoatAction:
			app.Action = &p
		case UsageWrapper:
			app.Usage = string(p)
		}
	}
	return app
}

type GoatApp struct {
	Name  string
	Usage string

	Action   *GoatAction
	Commands []GoatCommand
}

type Flag struct {
	Name     string
	Alias    string
	Usage    string
	Type     reflect.Type
	Required bool
}

func (f Flag) DisplayName() string {
	if f.Alias != "" {
		return f.Alias
	}
	return f.Name
}

func (f Flag) ArgName() string {
	return f.Name
}

type GoatAction struct {
	ActionValue reflect.Value
	ArgsType    reflect.Type
	Flags       []Flag
}

type GoatCommand interface {
	goatCommand()
}

type GoatCommandGroup struct {
	Name        string
	Usage       string
	Subcommands []GoatCommand
}
type GoatCommandSingle struct {
	Name   string
	Usage  string
	Action GoatAction
}

func (g GoatCommandGroup) goatCommand()  {}
func (g GoatCommandSingle) goatCommand() {}

func shouldEmbed(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.Struct
}
