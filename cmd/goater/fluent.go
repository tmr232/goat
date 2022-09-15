package main

import (
	"bytes"
	"github.com/pkg/errors"
	"go/ast"
	"go/format"
	"go/token"
)

type FluentCall struct {
	Name  string
	Args  []ast.Expr
	Ident *ast.Ident
}

type FluentChain struct {
	Base  ast.Expr
	Calls []FluentCall
}

func formatNode(fset *token.FileSet, node ast.Node) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		return "", errors.Wrap(err, "Failed formatting node")
	}
	return buf.String(), nil
}

func parseFluentChain(call *ast.CallExpr) FluentChain {
	var calls []FluentCall
	var selector *ast.SelectorExpr
	for {
		selector, _ = call.Fun.(*ast.SelectorExpr)
		args := call.Args
		name := selector.Sel.Name
		calls = append(calls, FluentCall{
			Name:  name,
			Args:  args,
			Ident: selector.Sel,
		})
		newCall, isCall := selector.X.(*ast.CallExpr)
		if !isCall {
			break
		}
		call = newCall
	}
	base := selector.X

	return FluentChain{
		Base:  base,
		Calls: Reversed(calls),
	}

}

type ActionDescription struct {
	Name  *string
	Usage *string
}

func isActionDescription(chain FluentChain) bool {
	base, isIdent := chain.Base.(*ast.Ident)
	if !isIdent {
		return false
	}
	if base.Name != "goat" {
		return false
	}
	if chain.Calls[0].Name != "Self" {
		return false
	}
	return true
}

func parseActionDescription(fset *token.FileSet, chain FluentChain, reportError func(ast.Node, string)) (ActionDescription, error) {
	description := ActionDescription{}

	for _, call := range chain.Calls[1:] {
		switch call.Name {
		case "Name":
			if description.Name != nil {
				reportError(call.Ident, "duplicate directive: .Name(name)")
				return ActionDescription{}, errors.New("Duplicate Name directive found")
			}
			if len(call.Args) != 1 {
				reportError(call.Ident, "Too many arguments passed to .Name(name)")
			}
			name, err := formatNode(fset, call.Args[0])
			if err != nil {
				reportError(call.Args[0], "Failed handling argument to .Name(name)")
				return ActionDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Name = &name

		case "Usage":
			if description.Usage != nil {
				reportError(call.Ident, "duplicate directive: .Usage(usage)")
				return ActionDescription{}, errors.New("Duplicate Usage directive found")
			}
			if len(call.Args) != 1 {
				reportError(call.Ident, "Too many arguments passed to .Usage(usage)")
			}
			usage, err := formatNode(fset, call.Args[0])
			if err != nil {
				reportError(call.Args[0], "Failed handling argument to .Usage(usage)")
				return ActionDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Usage = &usage

		default:
			reportError(call.Ident, "Unrecognized directive: "+call.Name)
			return ActionDescription{}, errors.New("unrecognized directive")
		}
	}

	return description, nil
}

type FlagDescription struct {
	Id      string
	Type    string
	Name    *string
	Usage   *string
	Default *string
}

func isFlagDescription(chain FluentChain) bool {
	base, isIdent := chain.Base.(*ast.Ident)
	if !isIdent {
		return false
	}
	if base.Name != "goat" {
		return false
	}
	if chain.Calls[0].Name != "Flag" {
		return false
	}
	return true
}

func parseFlagDescription(fset *token.FileSet, chain FluentChain, getType func(expr ast.Expr) (string, error), reportError func(ast.Node, string)) (FlagDescription, error) {
	id, err := formatNode(fset, chain.Calls[0].Args[0])
	if err != nil {
		return FlagDescription{}, errors.Wrap(err, "Could not format id node")
	}
	typ, err := getType(chain.Calls[0].Args[0])
	if err != nil {
		return FlagDescription{}, errors.Wrap(err, "Failed getting id type")
	}

	description := FlagDescription{Id: id, Type: typ}

	for _, call := range chain.Calls[1:] {
		switch call.Name {
		case "Name":
			if description.Name != nil {
				reportError(call.Ident, "duplicate directive: .Name(name)")
				return FlagDescription{}, errors.New("Duplicate Name directive found")
			}
			if len(call.Args) != 1 {
				reportError(call.Ident, "Too many arguments passed to .Name(name)")
			}
			name, err := formatNode(fset, call.Args[0])
			if err != nil {
				reportError(call.Args[0], "Failed handling argument to .Name(name)")
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Name = &name

		case "Usage":
			if description.Usage != nil {
				reportError(call.Ident, "duplicate directive: .Usage(usage)")
				return FlagDescription{}, errors.New("Duplicate Usage directive found")
			}
			if len(call.Args) != 1 {
				reportError(call.Ident, "Too many arguments passed to .Usage(usage)")
			}
			usage, err := formatNode(fset, call.Args[0])
			if err != nil {
				reportError(call.Args[0], "Failed handling argument to .Usage(usage)")
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Usage = &usage

		case "Default":
			if description.Default != nil {
				reportError(call.Ident, "duplicate directive: .Default(default_)")
				return FlagDescription{}, errors.New("Duplicate Default directive found")
			}
			if len(call.Args) != 1 {
				reportError(call.Ident, "Too many arguments passed to .Default(default_)")
			}
			default_, err := formatNode(fset, call.Args[0])
			if err != nil {
				reportError(call.Args[0], "Failed handling argument to .Default(default_)")
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Default = &default_
		default:
			reportError(call.Ident, "Unrecognized directive: "+call.Name)
			return FlagDescription{}, errors.New("unrecognized directive")
		}
	}

	return description, nil
}
