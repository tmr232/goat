package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/tmr232/goat"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"reflect"
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

type GoatData struct {
	Package string
	Apps    []GoatApp
}

type GoatArg struct {
	Name string
	Type string
}
type GoatSignature struct {
	Name string
	Args []GoatArg
}

func (gh *Goatherd) parseSignature(f *types.Func) (signature GoatSignature) {
	signature.Name = f.Name()

	funcSignature := f.Type().(*types.Signature)
	for i := 0; i < funcSignature.Params().Len(); i++ {
		param := funcSignature.Params().At(i)
		paramName := param.Name()
		paramType := param.Type().String()
		signature.Args = append(signature.Args, GoatArg{Name: paramName, Type: paramType})
	}
	return
}

type GoatDescription struct {
	Type string
	Flag string
}
type GoatApp struct {
	Func  string                     // Func is app function
	Flags map[string]GoatDescription // The flags for the app. Should later be a type that isn't CLI bound...
}

func (app GoatApp) ArgNames() (names []string) {
	for name, _ := range app.Flags {
		names = append(names, name)
	}
	return
}

func makeDefaultDescription(name, typ string) GoatDescription {
	switch typ {
	case "string":
		return GoatDescription{typ, fmt.Sprintf("%#v", goat.StringFlag{})}
	case "bool":
		return GoatDescription{typ, fmt.Sprintf("%#v", goat.BoolFlag{})}
	}
	log.Fatalf("Cannot describe type %s", typ)
	return GoatDescription{}
}

func MakeApp(signature GoatSignature, descriptions map[string]GoatDescription) (app GoatApp) {
	app.Func = signature.Name
	app.Flags = make(map[string]GoatDescription)
	for _, arg := range signature.Args {
		description, exists := descriptions[arg.Name]
		if exists {
			app.Flags[arg.Name] = description
		} else {
			app.Flags[arg.Name] = makeDefaultDescription(arg.Name, arg.Type)
		}
	}
	return
}

func DigFor[T any](base any, path ...string) (val T, found bool) {
	if base == nil {
		return *new(T), false
	}

	value := reflect.ValueOf(base)

	for _, pathPart := range path {
		for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			if value.IsNil() {
				return *new(T), false
			}
			value = value.Elem()
		}
		field := value.FieldByName(pathPart)
		if field == *new(reflect.Value) {
			return *new(T), false
		}
		value = field
	}

	result, isValid := value.Interface().(T)
	return result, isValid
}

func (gh *Goatherd) matchDescription(node ast.Node) bool {
	callExpr, isCallExpr := node.(*ast.CallExpr)
	if !isCallExpr {
		return false
	}

	pathToDescribeSelector := []string{"Fun", "X", "Fun"}
	pathToGoat := append(pathToDescribeSelector, "X")
	pathToDescribe := append(pathToDescribeSelector, "Sel")
	pathToAs := []string{"Fun", "Sel"}

	goatIdent, found := DigFor[*ast.Ident](callExpr, pathToGoat...)
	if !found {
		return false
	}
	describeIdent, found := DigFor[*ast.Ident](callExpr, pathToDescribe...)
	if !found {
		return false
	}
	asIdent, found := DigFor[*ast.Ident](callExpr, pathToAs...)
	if !found {
		return false
	}

	if goatIdent.Name != "goat" || describeIdent.Name != "Describe" || asIdent.Name != "As" {
		return false
	}

	return true
}

func (gh *Goatherd) parseDescription(node ast.Node) (string, GoatDescription) {
	asCall := node.(*ast.CallExpr)

	describeArgs, _ := DigFor[[]ast.Expr](asCall, "Fun", "X", "Args")
	ident := describeArgs[0].(*ast.Ident)
	name := ident.Name
	typ := gh.pkg.TypesInfo.TypeOf(ident).String()

	var out bytes.Buffer
	format.Node(&out, gh.pkg.Fset, asCall.Args[0])
	flag := out.String()

	return name, GoatDescription{Type: typ, Flag: flag}
}

func (gh *Goatherd) parseDescriptions(f *types.Func) map[string]GoatDescription {
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

	//var descriptions []*ast.CallExpr
	//ast.Inspect(fdecl.Body, func(node ast.Node) bool {
	//	if gh.isCallTo(node, "github.com/tmr232/goat", "Describe") {
	//		descriptions = append(descriptions, node.(*ast.CallExpr))
	//		var out bytes.Buffer
	//		format.Node(&out, gh.pkg.Fset, node)
	//		fmt.Println(out.String())
	//		return false
	//	}
	//	return true
	//})
	result := make(map[string]GoatDescription)
	ast.Inspect(fdecl.Body, func(node ast.Node) bool {
		if !gh.matchDescription(node) {
			return true
		}
		name, description := gh.parseDescription(node)
		result[name] = description
		return false
	})

	return result

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
		signature := gh.parseSignature(f)
		descriptions := gh.parseDescriptions(f)
		app := MakeApp(signature, descriptions)
		data.Apps = append(data.Apps, app)
		fmt.Printf("%#v\n", app)
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
