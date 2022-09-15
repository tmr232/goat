package main

import (
	"bytes"
	_ "embed"
	"github.com/pkg/errors"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
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

func (gh *Goatherd) isGoatRun(node *ast.CallExpr) bool {
	if len(node.Args) != 1 {
		return false
	}
	return gh.isCallTo(node, "github.com/tmr232/goat", "RunE") ||
		gh.isCallTo(node, "github.com/tmr232/goat", "Run")
}
func (gh *Goatherd) isGoatCommand(node *ast.CallExpr) bool {
	if len(node.Args) != 2 {
		return false
	}
	return gh.isCallTo(node, "github.com/tmr232/goat", "Command")
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
			if uses.Pkg() == nil {
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

func (gh *Goatherd) findActionFunctions() (actionFunctions []*types.Func) {
	callArgs := findActionCalls(gh)

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
		actionFunctions = append(actionFunctions, f)
	}
	return
}

func findNodesIf[T ast.Node](file *ast.File, pred func(node T) bool) []T {
	var matchingNodes []T
	for _, decl := range file.Decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			if typedNode, isRightType := node.(T); isRightType {
				if pred(typedNode) {
					matchingNodes = append(matchingNodes, typedNode)
					// We only stop recursion if we match the predicate
					return false
				}
			}
			return true
		})
	}
	return matchingNodes
}

func findActionCalls(gh *Goatherd) []ast.Expr {
	var callArgs []ast.Expr
	for _, syntax := range gh.pkg.Syntax {
		for _, call := range findNodesIf[*ast.CallExpr](syntax, gh.isGoatRun) {
			callArgs = append(callArgs, call.Args[0])
		}
		for _, call := range findNodesIf[*ast.CallExpr](syntax, gh.isGoatCommand) {
			callArgs = append(callArgs, call.Args[1])
		}
	}
	return callArgs
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

func (gh *Goatherd) parseArgDescription(callExpr *ast.CallExpr) (FlagDescription, bool) {
	chain := parseFluentChain(callExpr)
	description, err := parseFlagDesciption(gh.pkg.Fset, chain, func(expr ast.Expr) (string, error) {
		argType := gh.pkg.TypesInfo.TypeOf(expr)
		if argType == nil {
			return "", errors.New("Failed to find type of expression.")
		}
		return argType.String(), nil
	})
	if err != nil {
		log.Println(err)
		return FlagDescription{}, false
	}
	return description, true
}

func (gh *Goatherd) parseDescriptions(f *types.Func) []FlagDescription {
	fdecl := gh.findFuncDecl(f)

	var descriptions []FlagDescription
	ast.Inspect(fdecl.Body, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			// Keep going!
			return true
		}
		description, isOk := gh.parseArgDescription(callExpr)
		if !isOk {
			// Keep going
			return true
		}
		descriptions = append(descriptions, description)

		// Stop this branch
		return false
	})

	return descriptions

}

func (gh *Goatherd) findFuncDecl(f *types.Func) *ast.FuncDecl {
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
	return fdecl
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

type Flag struct {
	Type    string
	Name    string
	Usage   string
	Default string
}
type Action struct {
	Function string
	Flags    []Flag
}

func makeAction(signature GoatSignature, descriptions []FlagDescription) Action {
	flagByArgName := make(map[string]Flag)
	for _, arg := range signature.Args {
		flagByArgName[arg.Name] = Flag{
			Type:    arg.Type,
			Name:    "\"" + arg.Name + "\"",
			Default: "nil",
			Usage:   "\"\"",
		}
	}
	for _, desc := range descriptions {
		typ := desc.Type
		name := "\"" + desc.Id + "\""
		if desc.Name != nil {
			name = *desc.Name
		}
		usage := "\"\""
		if desc.Usage != nil {
			usage = *desc.Usage
		}
		default_ := "nil"
		if desc.Default != nil {
			default_ = *desc.Default
		}
		flagByArgName[desc.Id] = Flag{
			Type:    typ,
			Name:    name,
			Usage:   usage,
			Default: default_,
		}
	}
	var flags []Flag
	for _, arg := range signature.Args {
		flag := flagByArgName[arg.Name]
		flags = append(flags, flag)
	}
	return Action{
		Function: signature.Name,
		Flags:    flags,
	}
}

func main() {
	gh := NewGoatherd(loadPackages())
	var actions []Action
	for _, actionFunc := range gh.findActionFunctions() {
		signature := gh.parseSignature(actionFunc)
		descriptions := gh.parseDescriptions(actionFunc)
		actions = append(actions, makeAction(signature, descriptions))
	}
	data := struct {
		Package string
		Actions []Action
	}{
		Package: gh.pkg.Name,
		Actions: actions,
	}
	file, err := gh.Render("goat-file", data)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(gh.pkg.Name+"_goat.go", []byte(formatSource(file)), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}
