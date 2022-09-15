package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"testing"
)

func Test_parseFluentChain(t *testing.T) {
	code := "goat.A().B(1).C(\"hello\")"
	expr, err := parser.ParseExpr(code)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(parseFluentChain(expr.(*ast.CallExpr)))
}
