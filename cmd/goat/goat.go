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
	return gh.isCallTo(node, "github.com/tmr232/goat", "Run")
}

func (gh *Goatherd) isCallTo(node ast.Node, pkgPath, name string) bool {
	callExpr, isCall := node.(*ast.CallExpr)
	if !isCall {
		return false
	}

	found := false
	ast.Inspect(callExpr.Fun, func(node ast.Node) bool {
		if ident, isIdent := node.(*ast.Ident); isIdent {
			uses, exists := gh.pkg.TypesInfo.Uses[ident]
			if !exists {
				return false
			}
			if uses.Pkg().Path() == pkgPath && uses.Name() == name {
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
					if len(callExpr.Args) == 1 && gh.isGoatRun(callExpr) {
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
		definition, exists := gh.pkg.TypesInfo.Uses[ident]
		if !exists {
			log.Fatalf("%s goat.Run expects a function.", gh.pkg.Fset.Position(arg.Pos()))
		}
		f, isFunction := definition.(*types.Func)
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

func (gh *Goatherd) parseExtra(f *types.Func) {
	for i := 0; i < f.Scope().Len(); i++ {
		for _, n := range f.Scope().Names() {
			fmt.Println(n)
		}
	}

	var fdecl *ast.FuncDecl
	for _, syntax := range gh.pkg.Syntax {
		astObj := syntax.Scope.Lookup(f.Name())
		if astObj == nil {
			continue
		}
		decl, isFdecl := astObj.Decl.(*ast.FuncDecl)
		if !isFdecl {
			continue
		}
		if gh.pkg.TypesInfo.Defs[decl.Name] != f {
			continue
		}
		fdecl = decl
		break
	}
	if fdecl == nil {
		log.Fatal("Failed to find app")
	}

	var descriptions []*ast.CallExpr
	ast.Inspect(fdecl.Body, func(node ast.Node) bool {
		if gh.isCallTo(node, "github.com/tmr232/goat", "Describe") {
			descriptions = append(descriptions, node.(*ast.CallExpr))
			var out bytes.Buffer
			format.Node(&out, gh.pkg.Fset, node)
			fmt.Println(out.String())
			return false
		}
		return true
	})

	for _, desc := range descriptions {
		name := desc.Args[0].(*ast.Ident).Name
		fmt.Println("Name: ", name)
		info := desc.Args[1].(*ast.CompositeLit)
		for _, elt := range info.Elts {
			kv := elt.(*ast.KeyValueExpr)
			key := kv.Key.(*ast.Ident).Name
			var out bytes.Buffer
			format.Node(&out, gh.pkg.Fset, kv.Value)
			value := out.String()
			fmt.Println("Key: ", key, "Value: ", value)
		}
	}

	//var out bytes.Buffer
	//format.Node(&out, gh.pkg.Fset, fdecl)
	//fmt.Println(out.String())
	//ast.Print(gh.pkg.Fset, fdecl)
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
		gh.parseExtra(f)
	}
	file, err := gh.Render("goat-file", data)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(data.Package+"_goat.go", []byte(formatSource(file)), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}
