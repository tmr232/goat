package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

func Test_parseFluentDescription(t *testing.T) {
	code := "goat.Flag(value).Usage(\"Hello, World!\").Default(1).Name(\"v\",\"value\")"

	expr, err := parser.ParseExpr(code)
	fset := token.NewFileSet()
	if err != nil {
		return
	}
	if err != nil {
		t.Error(err)
	}
	chain := parseFluentChain(expr.(*ast.CallExpr))
	description, err := parseFluentDescription(fset, chain, func(node ast.Node) (string, error) { return "type", nil })
	if err != nil {
		t.Error(err)
	}
	fmt.Println(description)
}

//func Test_parseFluentChain(t *testing.T) {
//	type args struct {
//		call *ast.CallExpr
//	}
//	tests := []struct {
//		name string
//		args args
//		want FluentChain
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := parseFluentChain(tt.args.call); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("parseFluentChain() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
