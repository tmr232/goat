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

func Reversed[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, value := range slice {
		result[len(slice)-i-1] = value
	}
	return result
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

func isGoatFlag(chain FluentChain) bool {
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

type FluentDescription struct {
	Name        string
	Type        string
	Descriptors map[string][]string
}

func Map[T any, R any](slice []T, op func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = op(v)
	}
	return result
}

func MapE[T any, R any](slice []T, op func(T) (R, error)) ([]R, error) {
	result := make([]R, len(slice))
	var err error
	for i, v := range slice {
		result[i], err = op(v)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("op %v failed to map over value: %v", &op, v))
		}
	}
	return result, nil
}

func parseFluentDescription(fset *token.FileSet, chain FluentChain, getType func(ast.Node) (string, error)) (FluentDescription, error) {
	if !isGoatFlag(chain) {
		return FluentDescription{}, errors.New("Not a goat flag description!")
	}

	name, err := formatNode(fset, chain.Calls[0].Args[0])
	if err != nil {
		return FluentDescription{}, errors.Wrap(err, "Could not format name node")
	}
	typ, err := getType(chain.Calls[0].Args[0])
	if err != nil {
		return FluentDescription{}, errors.Wrap(err, "Failed getting name type")
	}

	descriptors := make(map[string][]string, len(chain.Calls)-1)
	for _, call := range chain.Calls[1:] {
		args, err := MapE(call.Args, func(node ast.Expr) (string, error) { return formatNode(fset, node) })
		if err != nil {
			return FluentDescription{}, errors.Wrap(err, "Failed converting args to strings")
		}
		descriptors[call.Name] = args
	}

	return FluentDescription{
		Name:        name,
		Type:        typ,
		Descriptors: descriptors,
	}, nil
}
