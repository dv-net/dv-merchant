package dbtxcheck

import (
	"go/ast"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("dbtxcheck", New)
}

type DBTxCheckPlugin struct{}

func New(_ any) (register.LinterPlugin, error) {
	return &DBTxCheckPlugin{}, nil
}

func (p *DBTxCheckPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "dbtxcheck",
			Doc:  "disallow usage of pgx.BeginTxFunc",
			Run:  p.run,
		},
	}, nil
}

func (p *DBTxCheckPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func (p *DBTxCheckPlugin) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		pgxImportName := findPgxImportName(file)
		if pgxImportName == "" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == pgxImportName && sel.Sel.Name == "BeginTxFunc" {
				pass.Report(analysis.Diagnostic{
					Pos:     call.Pos(),
					Message: "usage of pgx.BeginTxFunc is prohibited - use repo tx wrapper instead",
				})
			}
			return true
		})
	}

	return nil, nil
}

func findPgxImportName(file *ast.File) string {
	for _, imp := range file.Imports {
		if strings.Contains(imp.Path.Value, "github.com/jackc/pgx/") {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return "pgx"
		}
	}

	return ""
}
