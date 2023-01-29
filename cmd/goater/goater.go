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
		Mode:       packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax | packages.NeedName | packages.NeedImports | packages.NeedDeps,
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

type actionDefinition struct {
	Func *types.Func
	Def  ast.Node
}

func (gh *Goatherd) findActionFunctions() (actionFunctions []actionDefinition) {
	callArgs := findActionCalls(gh)

	// For the time being - only ast.Ident nodes will be considered valid because
	// we need their base definition.

	for _, arg := range callArgs {
		var ident *ast.Ident
		switch node := arg.(type) {
		case *ast.Ident:
			ident = node
		case *ast.SelectorExpr:
			x, isIdent := node.X.(*ast.Ident)
			if isIdent {
				obj := gh.pkg.TypesInfo.ObjectOf(x)
				if obj != nil {
					fmt.Println(reflect.TypeOf(obj))
					_, isPkg := obj.(*types.PkgName)
					if isPkg {
						fmt.Println("yay!")
					}
				}
			}
			ident = node.Sel
		default:
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
		actionFunctions = append(actionFunctions, actionDefinition{Func: f, Def: arg})
	}
	return actionFunctions
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
	NoError bool
	Func    *types.Func
}

func printTypeInfo(t types.Type) {
	fmt.Println("================")
	fmt.Println(t)
	fmt.Println(reflect.TypeOf(t))
	switch t := t.(type) {
	case *types.Pointer:
		printTypeInfo(t.Elem())
	case *types.Named:
		fmt.Println(t.Obj())
		fmt.Println(t.Obj().Pkg(), t.Obj().Name())
	}
}

func getParams(f *types.Func) []*types.Var {
	params := f.Type().(*types.Signature).Params()
	var result []*types.Var
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		result = append(result, param)
	}
	return result
}

func returnsError(f *types.Func) (bool, error) {
	funcSignature := f.Type().(*types.Signature)

	results := funcSignature.Results()
	switch results.Len() {
	case 0:
		return true, nil
	case 1:
		result := results.At(0)
		if result.Type().String() != "error" {
			return false, errors.New("Action function result must be either error or nothing")
		}
		return false, nil
	default:
		return false, errors.New("Action function returns more than 1 value")
	}
}

func (gh *Goatherd) parseSignature(f *types.Func) (signature GoatSignature, err error) {
	signature.Func = f

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

	return signature, nil
}

var notAFlagDescription = errors.New("Not a flag description")

func (gh *Goatherd) reportError(node ast.Node, message string) {
	fmt.Println(gh.pkg.Fset.Position(node.Pos()), "Error:", message)
}
func (gh *Goatherd) parseActionDescription(fdecl *ast.FuncDecl) (ActionDescription, error) {
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

func (gh *Goatherd) parseFlagDescriptions(fdecl *ast.FuncDecl) ([]FlagDescription, error) {
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
	// Weird hack to get the package containing the function
	pkg := gh.pkg
	if pkg.PkgPath != f.Pkg().Path() {
		var ok bool
		pkg, ok = pkg.Imports[f.Pkg().Path()]
		if !ok {
			log.Fatalf("Failed finding package %s", f.Pkg().Path())
		}
	}
	for _, syntax := range pkg.Syntax {
		astObj := syntax.Scope.Lookup(f.Name())
		if astObj == nil {
			continue
		}
		decl, isFdecl := astObj.Decl.(*ast.FuncDecl)
		if !isFdecl {
			continue
		}
		if pkg.TypesInfo.Defs[decl.Name] != f {
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

type ImportManager struct {
	thisPkgPath   string
	importsByPath map[string]string
	importsByName map[string]string
}

func NewImportManager(thisPkgPath string) *ImportManager {
	return &ImportManager{
		thisPkgPath:   thisPkgPath,
		importsByPath: make(map[string]string),
		importsByName: make(map[string]string),
	}
}

func (im *ImportManager) getImports() []string {
	// TODO: Remove aliases where possible
	var imports []string
	for alias, path := range im.importsByName {
		imports = append(imports, fmt.Sprintf("%s \"%s\"", alias, path))
	}
	return imports
}

func (im *ImportManager) addImport(name, path string) (alias string) {
	// If we already have an alias for this path - we just return it
	if alias, exists := im.importsByPath[path]; exists {
		return alias
	}
	// If the desired name already exists - we create an alternate alias
	if _, exists := im.importsByName[name]; exists {
		suffix := 0
		for {
			alias := fmt.Sprintf("%s_%d", name, suffix)
			if _, exists := im.importsByName[alias]; !exists {
				return im.addImport(alias, path)
			}
		}
	}
	// If this is a new name and a new path - we add it
	im.importsByName[name] = path
	im.importsByPath[path] = name

	return name
}

func (im *ImportManager) getTypeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg.Path() == im.thisPkgPath {
			return t.Obj().Name()
		}
		alias := im.addImport(pkg.Name(), pkg.Path())
		return alias + "." + t.Obj().Name()
	case *types.Pointer:
		return "*" + im.getTypeName(t.Elem())
	case *types.Basic:
		return t.Name()
	default:
		panic("WTF is this type???")
	}
}

func makeAction(imports *ImportManager, functionName string, signature GoatSignature, actionDescription ActionDescription, flagDescriptions []FlagDescription) Action {
	flagByArgName := make(map[string]Flag)
	for _, param := range getParams(signature.Func) {
		name := imports.getTypeName(param.Type())
		flagByArgName[param.Name()] = Flag{
			//Type:      param.Type().String(),
			Type:      name,
			Name:      fmt.Sprintf("\"%s\"", param.Name()),
			Default:   "nil",
			Usage:     "\"\"",
			IsContext: isGoatContext(param.Type().String()),
		}
	}
	for _, desc := range flagDescriptions {

		flag := flagByArgName[desc.Id]

		if desc.Name != nil {
			flag.Name = *desc.Name
		}
		if desc.Usage != nil {
			flag.Usage = *desc.Usage
		}
		if desc.Default != nil {
			flag.Default = *desc.Default
		}

		flagByArgName[desc.Id] = flag
	}
	var flags []Flag
	for _, param := range getParams(signature.Func) {
		flag := flagByArgName[param.Name()]
		flags = append(flags, flag)
	}

	name := "\"" + signature.Func.Name() + "\""
	if actionDescription.Name != nil {
		name = *actionDescription.Name
	}
	usage := "\"\""
	if actionDescription.Usage != nil {
		usage = *actionDescription.Usage
	}

	return Action{
		Function: functionName,
		Flags:    flags,
		Name:     name,
		Usage:    usage,
		NoError:  signature.NoError,
	}
}

func (gh *Goatherd) getFuncDecl(f *types.Func) *ast.FuncDecl {
	for _, file := range gh.pkg.Syntax {
		for _, decl := range file.Decls {
			funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
			if !isFuncDecl {
				continue
			}
			if gh.pkg.TypesInfo.Defs[funcDecl.Name] == f {
				return funcDecl
			}
		}
	}
	return nil
}

type ImportByName map[string]*ast.ImportSpec

func (gh *Goatherd) getImports() ImportByName {
	importsByName := make(ImportByName, 0)
	for _, importSpec := range gh.pkg.Syntax[0].Imports {
		var name string
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		} else {
			p := gh.pkg.Imports[strings.Trim(importSpec.Path.Value, "\"")]
			name = p.Name
		}
		importsByName[name] = importSpec
	}

	return importsByName
}

func (gh *Goatherd) createAction(imports *ImportManager, actionFunc actionDefinition) (Action, error) {
	// The AST declaration is used in multiple places, so we get it here.
	fdecl := gh.findFuncDecl(actionFunc.Func)
	signature, err := gh.parseSignature(actionFunc.Func)
	if err != nil {
		log.Fatal(err)
	}
	actionDescription, err := gh.parseActionDescription(fdecl)
	if err != nil {
		log.Fatal(err)
	}
	flagDescriptions, err := gh.parseFlagDescriptions(fdecl)
	if err != nil {
		log.Fatal(err)
	}
	functionName, err := formatNode(gh.pkg.Fset, actionFunc.Def)
	if err != nil {
		log.Fatal(err)
	}
	return makeAction(imports, functionName, signature, actionDescription, flagDescriptions), nil
}

func main() {
	gh := NewGoatherd(loadPackages())
	var actions []Action
	//usedImports := make(map[string]bool)
	importManager := NewImportManager(gh.pkg.PkgPath)
	for name, path := range map[string]string{
		"goat":  "github.com/tmr232/goat",
		"flags": "github.com/tmr232/goat/flags",
		"cli":   "github.com/urfave/cli/v2",
	} {
		importManager.addImport(name, path)
	}
	for _, actionFuncDefinition := range gh.findActionFunctions() {
		//if selector, isSelector := actionFuncDefinition.Def.(*ast.SelectorExpr); isSelector {
		//	usedImports[selector.X.(*ast.Ident).Name] = true
		//
		//}
		action, err := gh.createAction(importManager, actionFuncDefinition)
		if err != nil {
			log.Fatal(err)
		}
		actions = append(actions, action)
	}
	//
	//importsByPath := make(map[string]*string)
	//for name, spec := range gh.getImports() {
	//	if !usedImports[name] {
	//		continue
	//	}
	//	var alias *string
	//	if spec.Name != nil {
	//		alias = &spec.Name.Name
	//	}
	//	importsByPath[spec.Path.Value] = alias
	//}
	//
	//baseImports := []string{
	//	"\"github.com/tmr232/goat\"",
	//	"\"github.com/tmr232/goat/flags\"",
	//	"\"github.com/urfave/cli/v2\"",
	//}
	//
	//imports := append([]string{}, baseImports...)
	//for path, name := range importsByPath {
	//	if name == nil {
	//		imports = append(imports, path)
	//	} else {
	//		imports = append(imports, *name+" "+path)
	//	}
	//}

	imports := importManager.getImports()

	data := struct {
		Package string
		Actions []Action
		Imports []string
	}{
		Package: gh.pkg.Name,
		Actions: actions,
		Imports: imports,
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
