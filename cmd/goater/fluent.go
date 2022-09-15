package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/format"
	"go/token"
)

type FluentCall struct {
	Name string
	Args []ast.Expr
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
		fmt.Println(name, args)
		calls = append(calls, FluentCall{
			Name: name,
			Args: args,
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

func parseFlagDesciption(fset *token.FileSet, chain FluentChain, getType func(expr ast.Expr) (string, error)) (FlagDescription, error) {
	if !isFlagDescription(chain) {
		return FlagDescription{}, errors.New("Not a goat flag description!")
	}

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
				return FlagDescription{}, errors.New("Duplicate name directive found")
			}
			name, err := formatNode(fset, call.Args[0])
			if err != nil {
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Name = &name

		case "Usage":
			if description.Usage != nil {
				return FlagDescription{}, errors.New("Duplicate usage directive found")
			}
			name, err := formatNode(fset, call.Args[0])
			if err != nil {
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Usage = &name

		case "Default":
			if description.Default != nil {
				return FlagDescription{}, errors.New("Duplicate default directive found")
			}
			name, err := formatNode(fset, call.Args[0])
			if err != nil {
				return FlagDescription{}, errors.Wrap(err, "Failed formatting argument")
			}
			description.Default = &name
		}
	}

	return description, nil
}
