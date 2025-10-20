package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	panicMsg    = "использование встроенной функции panic запрещено"
	logFatalMsg = "log.Fatal допустим только в функции main пакета main"
	osExitMsg   = "os.Exit допустим только в функции main пакета main"
)

var Analyzer = &analysis.Analyzer{
	Name: "metricslinter",
	Doc:  "проверяет использование panic, log.Fatal и os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			inspectFunc(pass, pkgName, fn)
		}
	}

	return nil, nil
}

func inspectFunc(pass *analysis.Pass, pkgName string, fn *ast.FuncDecl) {
	isMainFunc := pkgName == "main" && fn.Recv == nil && fn.Name != nil && fn.Name.Name == "main"

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if ident, ok := call.Fun.(*ast.Ident); ok {
			if ident.Name == "panic" && ident.Obj == nil {
				pass.Reportf(call.Pos(), panicMsg)
				return true
			}
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		pkgNameObj, ok := pass.TypesInfo.Uses[pkgIdent].(*types.PkgName)
		if !ok {
			return true
		}

		imported := pkgNameObj.Imported()
		switch imported.Path() {
		case "log":
			if sel.Sel.Name == "Fatal" && !isMainFunc {
				reportCall(pass, call, logFatalMsg)
			}
		case "os":
			if sel.Sel.Name == "Exit" && !isMainFunc {
				reportCall(pass, call, osExitMsg)
			}
		}

		return true
	})
}

func reportCall(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	pos := call.Pos()
	if call.Lparen.IsValid() {
		pos = call.Lparen
	}
	pass.Reportf(pos, msg)
}
