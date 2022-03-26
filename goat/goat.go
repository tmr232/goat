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

func Command[Args any](name string, action func(Args) error, parts ...CommandPart) GoatCommandSingle {
	cmd := GoatCommandSingle{
		Name: name,
	}

	var args ArgMap
	for _, p := range parts {
		switch p := p.(type) {
		case UsageWrapper:
			cmd.Usage = string(p)
		case ArgMap:
			args = p
		}
	}

	cmd.Action = actionWithArgs(action, args)

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

func Flags(argsType reflect.Type, args ArgMap) (flags []Flag) {
	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)

		if shouldEmbed(field.Type) {
			flags = append(flags, Flags(field.Type, args)...)
		} else {
			name := field.Name

			var alias string
			var usage string

			arg, hasArg := args[name]
			if hasArg {
				alias = arg.Alias
				usage = arg.Usage
			} else {
				alias = field.Tag.Get("alias")
				usage = field.Tag.Get("usage")
			}
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
	return actionWithArgs(action, nil)
}

func actionWithArgs[Args any](action func(Args) error, args ArgMap) GoatAction {
	return GoatAction{
		ActionValue: reflect.ValueOf(action),
		ArgsType:    reflect.TypeOf(action).In(0),
		Flags:       Flags(reflect.TypeOf(action).In(0), args),
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

type Arg struct {
	Name  string
	Alias string
	Usage string
}

type ArgMap map[string]Arg

type CommandPart interface {
	commandPart()
}

func (a ArgMap) commandPart()       {}
func (u UsageWrapper) commandPart() {}

func With(args ...Arg) ArgMap {
	argMap := make(map[string]Arg)
	for _, arg := range args {
		argMap[arg.Name] = arg
	}
	return argMap
}
