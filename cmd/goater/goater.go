package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tmr232/goat"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
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

func (gh *Goatherd) isGoatRun(node *ast.CallExpr) bool {
	if len(node.Args) != 1 {
		return false
	}
	return gh.isCallTo(node, "github.com/tmr232/goat", "RunE") ||
		gh.isCallTo(node, "github.com/tmr232/goat", "Run")
}
func (gh *Goatherd) isGoatCommand(node *ast.CallExpr) bool {
	if len(node.Args) < 1 {
		return false
	}
	return gh.isCallTo(node, "github.com/tmr232/goat", "Command")
}

type callTarget struct {
	PkgPath string
	Name    string
}

func isCallTo(target callTarget, node ast.Node, typesInfo *types.Info) bool {
	_, isCall := node.(*ast.CallExpr)
	if !isCall {
		return false
	}

	for {
		switch current := node.(type) {
		case *ast.CallExpr:
			node = current.Fun
		case *ast.SelectorExpr:
			node = current.Sel
		case *ast.Ident:
			definition, exists := typesInfo.Uses[current]
			if !exists {
				return false
			}

			funcDef, isFunc := definition.(*types.Func)
			if !isFunc {
				return false
			}

			if funcDef.Pkg() == nil {
				return false
			}
			if funcDef.Pkg().Path() == target.PkgPath && funcDef.Name() == target.Name {
				return true
			}
			return false
		default:
			return false
		}
	}
}

func (gh *Goatherd) isCallTo(node ast.Node, pkgPath, name string) bool {
	return isCallTo(callTarget{Name: name, PkgPath: pkgPath}, node, gh.pkg.TypesInfo)
}

func (gh *Goatherd) findActionFunctions() (actionFunctions []*types.Func) {
	callArgs := findActionCalls(gh)

	// For the time being - only ast.Ident nodes will be considered valid because
	// we need their base definition.

	for _, arg := range callArgs {
		ident, isIdent := arg.(*ast.Ident)
		if !isIdent {
			//funcLit := arg.(*ast.FuncLit)
			//fmt.Println(gh.pkg.TypesInfo.TypeOf(funcLit))
			//fmt.Println(reflect.TypeOf(gh.pkg.TypesInfo.TypeOf(funcLit)))
			//fmt.Println(gh.pkg.TypesInfo.TypeOf(funcLit).(*types.Signature))
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
					// We recurse the entire AST without stopping as there may be
					// nested calls when we create subcommands.
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
			callArgs = append(callArgs, call.Args[0])
		}
	}
	return callArgs
}

type GoatArg struct {
	Name string
	Type string
}
type GoatSignature struct {
	Name    string
	Args    []GoatArg
	NoError bool
}

func (gh *Goatherd) parseSignature(f *types.Func) (signature GoatSignature, err error) {
	signature.Name = f.Name()

	funcSignature := f.Type().(*types.Signature)

	results := funcSignature.Results()
	switch results.Len() {
	case 0:
		signature.NoError = true
	case 1:
		signature.NoError = false
		result := results.At(0)
		if result.Type().String() != "error" {
			return GoatSignature{}, errors.New("Action function result must be either error or nothing")
		}
	default:
		return GoatSignature{}, errors.New("Action function returns more than 1 value")
	}

	for i := 0; i < funcSignature.Params().Len(); i++ {
		param := funcSignature.Params().At(i)
		paramName := param.Name()
		paramType := param.Type().String()
		signature.Args = append(signature.Args, GoatArg{Name: paramName, Type: paramType})
	}
	return signature, nil
}

var notAFlagDescription = errors.New("Not a flag description")

func (gh *Goatherd) reportError(node ast.Node, message string) {
	fmt.Println(gh.pkg.Fset.Position(node.Pos()), "Error:", message)
}
func (gh *Goatherd) parseActionDescription(f *types.Func) (ActionDescription, error) {
	fdecl := gh.findFuncDecl(f)

	var description ActionDescription
	var err error
	ast.Inspect(fdecl.Body, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			// Keep going!
			return true
		}
		chain, isChain := parseFluentChain(callExpr)
		if !isChain || !isActionDescription(chain) {
			// Keep going
			return true
		}
		description, err = parseActionDescription(gh.pkg.Fset, chain, gh.reportError)
		// Stop this branch
		return false
	})

	if err != nil {
		return ActionDescription{}, errors.Wrap(err, "Failed parsing ActionDescription")
	}

	if description.Usage == nil && fdecl.Doc != nil {
		doc := fdecl.Doc.Text()
		doc = strings.TrimPrefix(doc, fdecl.Name.Name)
		text := fmt.Sprintf("%#v", strings.TrimSpace(doc))
		description.Usage = &text
	}
	return description, nil

}

func (gh *Goatherd) parseFlagDescriptions(f *types.Func) ([]FlagDescription, error) {
	fdecl := gh.findFuncDecl(f)

	var parseErrors []error

	var descriptions []FlagDescription
	ast.Inspect(fdecl.Body, func(node ast.Node) bool {
		callExpr, isCall := node.(*ast.CallExpr)
		if !isCall {
			// Keep going!
			return true
		}
		description, err := gh.parseFlagDescription(callExpr)
		if err == notAFlagDescription {
			// Keep going
			return true
		}
		if err != nil {
			parseErrors = append(parseErrors, err)
		}
		descriptions = append(descriptions, description)

		// Stop this branch
		return false
	})

	if len(parseErrors) != 0 {
		return nil, errors.New("Encountered errors!")
	}
	return descriptions, nil

}
func (gh *Goatherd) parseFlagDescription(callExpr *ast.CallExpr) (FlagDescription, error) {
	chain, isChain := parseFluentChain(callExpr)
	if !isChain || !isFlagDescription(chain) {
		return FlagDescription{}, notAFlagDescription
	}
	description, err := parseFlagDescription(gh.pkg.Fset, chain, func(expr ast.Expr) (string, error) {
		argType := gh.pkg.TypesInfo.TypeOf(expr)
		if argType == nil {
			return "", errors.New("Failed to find type of expression.")
		}
		return argType.String(), nil
	},
		gh.reportError)
	if err != nil {
		return FlagDescription{}, err
	}
	return description, nil
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
	Type      string
	Name      string
	Usage     string
	Default   string
	IsContext bool
}
type Action struct {
	Function string
	Flags    []Flag
	Name     string
	Usage    string
	NoError  bool
}

func isGoatContext(typeName string) bool {
	goatContextType := reflect.TypeOf(*new(goat.Context))
	goatContextTypeName := "*" + goatContextType.PkgPath() + "." + goatContextType.Name()
	return typeName == goatContextTypeName
}

func makeAction(signature GoatSignature, actionDescription ActionDescription, flagDescriptions []FlagDescription) Action {
	flagByArgName := make(map[string]Flag)
	for _, arg := range signature.Args {
		flagByArgName[arg.Name] = Flag{
			Type:      arg.Type,
			Name:      "\"" + arg.Name + "\"",
			Default:   "nil",
			Usage:     "\"\"",
			IsContext: isGoatContext(arg.Type),
		}
	}
	for _, desc := range flagDescriptions {
		typ := desc.Type
		if isGoatContext(typ) {
			// goat.Context is not a flag type, it's always the context object.
			continue
		}
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
			Type:      typ,
			Name:      name,
			Usage:     usage,
			Default:   default_,
			IsContext: false,
		}
	}
	var flags []Flag
	for _, arg := range signature.Args {
		flag := flagByArgName[arg.Name]
		flags = append(flags, flag)
	}

	name := "\"" + signature.Name + "\""
	if actionDescription.Name != nil {
		name = *actionDescription.Name
	}
	usage := "\"\""
	if actionDescription.Usage != nil {
		usage = *actionDescription.Usage
	}

	return Action{
		Function: signature.Name,
		Flags:    flags,
		Name:     name,
		Usage:    usage,
		NoError:  signature.NoError,
	}
}

func (gh *Goatherd) getFunctionComment(f *types.Func) *string {
	for _, file := range gh.pkg.Syntax {
		for _, decl := range file.Decls {
			funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
			if !isFuncDecl {
				continue
			}
			if gh.pkg.TypesInfo.Defs[funcDecl.Name] == f {
				if funcDecl.Doc == nil {
					return nil
				}
				text := funcDecl.Doc.Text()
				return &text
			}
		}
	}
	return nil
}

func main() {
	gh := NewGoatherd(loadPackages())
	var actions []Action
	for _, actionFunc := range gh.findActionFunctions() {
		signature, err := gh.parseSignature(actionFunc)
		if err != nil {
			log.Fatal(err)
		}
		actionDescription, err := gh.parseActionDescription(actionFunc)
		if err != nil {
			log.Fatal(err)
		}
		flagDescriptions, err := gh.parseFlagDescriptions(actionFunc)
		if err != nil {
			log.Fatal(err)
		}
		actions = append(actions, makeAction(signature, actionDescription, flagDescriptions))
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
