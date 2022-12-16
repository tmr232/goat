package goat

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"io"
	"reflect"
)

type Context struct {
	cliContext    *cli.Context
	flagsByAction map[reflect.Value]map[string]any
}

func (ctx *Context) GetFlag(f any, name string) (any, error) {
	actionFlags, exists := ctx.flagsByAction[reflect.ValueOf(f)]
	if !exists {
		return nil, errors.New("Action wasn't triggered")
	}
	flag, exists := actionFlags[name]
	if !exists {
		return nil, errors.New("Flag doesn't exist")
	}
	return flag, nil
}

func (ctx *Context) GetWriter() io.Writer {
	return ctx.cliContext.App.Writer
}

func GetFlag[T any](ctx *Context, f any, name string) (T, error) {
	anyVal, err := ctx.GetFlag(f, name)
	if err != nil {
		return *new(T), err
	}
	// We don't check for errors here because a bad cast here is a programmer error.
	return anyVal.(T), nil
}

func GetContext(c *cli.Context) *Context {
	flagsByAction := make(map[reflect.Value]map[string]any)
	for _, parentCtx := range c.Lineage()[1:] {
		if parentCtx.App == nil {
			break
		}
		funcValue, isRegistered := functionByCliActionFunc[reflect.ValueOf(parentCtx.App.Action)]
		if !isRegistered {
			break
		}
		ctxBuilder := runConfigByFunction[funcValue].CtxFlagBuilder
		flagsByAction[funcValue] = ctxBuilder(c)
	}

	return &Context{flagsByAction: flagsByAction, cliContext: c}
}
