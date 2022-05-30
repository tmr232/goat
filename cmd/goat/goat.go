package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

func loadPackages() *packages.Package {
	cfg := &packages.Config{
		Mode:       packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax | packages.NeedName | packages.NeedImports,
		Context:    nil,
		Logf:       nil,
		Dir:        "",
		Env:        nil,
		BuildFlags: nil,
		Fset:       nil,
		ParseFile:  nil,
		Tests:      false,
		Overlay:    nil,
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("Expected 1 package, found %d", len(pkgs))
	}

	return pkgs[0]
}

//go:embed goat.tmpl
var coreTemplate string

type Goatherd struct {
	pkg      *packages.Package
	template *template.Template
}

func NewGoatherd(pkg *packages.Package) *Goatherd {
	funcMap := template.FuncMap{
		"join":       strings.Join,
		"trimPrefix": strings.TrimPrefix,
	}
	t, err := template.New("core").Funcs(funcMap).Parse(coreTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return &Goatherd{pkg: pkg, template: t}
}

func (gh *Goatherd) Render(name string, data any) (string, error) {
	var out bytes.Buffer
	err := gh.template.ExecuteTemplate(&out, name, data)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (gh *Goatherd) isGoatRun(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		if ident, isIdent := node.(*ast.Ident); isIdent {
			uses, exists := gh.pkg.TypesInfo.Uses[ident]
			if !exists {
				return false
			}
			if uses.Pkg().Path() == "github.com/tmr232/goat" && uses.Name() == "Run" {
				found = true
			}
		}
		return !found
	})
	return found
}

func (gh *Goatherd) findGoatApps() (apps []*types.Func) {
	var callArgs []ast.Expr
	for _, syntax := range gh.pkg.Syntax {
		for _, decl := range syntax.Decls {
			ast.Inspect(decl, func(node ast.Node) bool {
				if callExpr, isCall := node.(*ast.CallExpr); isCall {
					if len(callExpr.Args) == 1 && gh.isGoatRun(callExpr.Fun) {
						callArgs = append(callArgs, callExpr.Args[0])
					}
					return false
				}
				return true
			})
		}
	}

	// For the time being - only ast.Ident nodes will be considered valid because
	// we need their base definition.

	for _, arg := range callArgs {
		ident, isIdent := arg.(*ast.Ident)
		if !isIdent {
			log.Fatalf("%s goat.Run only accepts free functions.", gh.pkg.Fset.Position(arg.Pos()))
		}
		defitition, exists := gh.pkg.TypesInfo.Uses[ident]
		if !exists {
			log.Fatalf("%s goat.Run expects a function.", gh.pkg.Fset.Position(arg.Pos()))
		}
		f, isFunction := defitition.(*types.Func)
		if !isFunction {
			log.Fatalf("%s goat.Run expects a function.", gh.pkg.Fset.Position(arg.Pos()))
		}
		apps = append(apps, f)
	}
	return
}

type Arg struct {
	Name string
	Type string
	gh   *Goatherd
}

func (a Arg) AsArg() string {
	return fmt.Sprintf("%s %s", a.Name, a.Type)
}

func (a Arg) AsFlag() string {
	switch a.Type {
	case "string":
		flag, err := a.gh.Render("string-flag", a)
		if err != nil {
			log.Fatal(err)
		}
		return flag
	case "bool":
		flag, err := a.gh.Render("bool-flag", a)
		if err != nil {
			log.Fatal(err)
		}
		return flag
	}
	log.Fatalf("Unsupported type %s", a.Type)
	return ""
}

type Args []Arg

func (args Args) Names() (names []string) {
	for _, arg := range args {
		names = append(names, arg.Name)
	}
	return
}

type App struct {
	Func string
	Args Args
}

type GoatData struct {
	Package string
	Apps    []App
}

func (gh *Goatherd) parseApp(f *types.Func) (app App) {
	app.Func = f.Name()

	signature := f.Type().(*types.Signature)
	for i := 0; i < signature.Params().Len(); i++ {
		param := signature.Params().At(i)
		paramName := param.Name()
		paramType := param.Type().String()
		app.Args = append(app.Args, Arg{Name: paramName, Type: paramType, gh: gh})
	}
	return
}

func formatSource(src string) string {
	formattedSrc, err := format.Source([]byte(src))
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return src
	}
	return string(formattedSrc)
}

func main() {
	gh := NewGoatherd(loadPackages())
	data := GoatData{
		Package: gh.pkg.Name,
	}
	for _, f := range gh.findGoatApps() {
		data.Apps = append(data.Apps, gh.parseApp(f))
	}
	file, err := gh.Render("goat-file", data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(formatSource(file))

	err = ioutil.WriteFile(data.Package+"_goat.go", []byte(formatSource(file)), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}
