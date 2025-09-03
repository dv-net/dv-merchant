package structinit

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("structinit", New)
}

type StructInitPlugin struct{}

func New(_ any) (register.LinterPlugin, error) {
	return &StructInitPlugin{}, nil
}

func (p *StructInitPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "structinit",
			Doc:  "check for uninitialized struct fields in constructors",
			Run:  p.run,
		},
	}, nil
}

func (p *StructInitPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func (p *StructInitPlugin) run(pass *analysis.Pass) (any, error) {
	// Find all struct types and their fields
	structTypes := make(map[*types.Named]map[string]*types.Var)
	// Track field usage across all methods
	fieldUsage := make(map[*types.Named]map[string]bool)
	// Track dynamic field assignments
	fieldAssignments := make(map[*types.Named]map[string]bool)
	// Track nil-safe field usage (fields checked for nil before use)
	nilSafeFields := make(map[*types.Named]map[string]bool)

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.GenDecl:
				if x.Tok == token.TYPE {
					for _, spec := range x.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := ts.Type.(*ast.StructType); ok {
								// Get the type info for this struct
								obj := pass.TypesInfo.Defs[ts.Name]
								if obj != nil {
									if named, ok := obj.Type().(*types.Named); ok {
										fields := make(map[string]*types.Var)
										if structType, ok := named.Underlying().(*types.Struct); ok {
											for i := 0; i < structType.NumFields(); i++ {
												field := structType.Field(i)
												if !field.Embedded() {
													fields[field.Name()] = field
												}
											}
											structTypes[named] = fields
											fieldUsage[named] = make(map[string]bool)
											fieldAssignments[named] = make(map[string]bool)
											nilSafeFields[named] = make(map[string]bool)
										}
									}
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				// Check constructor functions
				if x.Name != nil && strings.HasPrefix(x.Name.Name, "New") {
					p.checkConstructor(pass, x, structTypes, fieldUsage, fieldAssignments, nilSafeFields)
				}
				// Check all methods for field usage and assignments
				if x.Recv != nil && len(x.Recv.List) > 0 {
					p.analyzeMethodFieldUsage(pass, x, structTypes, fieldUsage, fieldAssignments, nilSafeFields)
				}
			}
			return true
		})
	}

	return nil, nil
}

func (p *StructInitPlugin) analyzeMethodFieldUsage(pass *analysis.Pass, fn *ast.FuncDecl, structTypes map[*types.Named]map[string]*types.Var, fieldUsage map[*types.Named]map[string]bool, fieldAssignments map[*types.Named]map[string]bool, nilSafeFields map[*types.Named]map[string]bool) {
	if fn.Body == nil || fn.Recv == nil || len(fn.Recv.List) == 0 {
		return
	}

	// Get the receiver type
	recv := fn.Recv.List[0]
	var structType *types.Named

	if recvType := pass.TypesInfo.Types[recv.Type]; recvType.Type != nil {
		switch t := recvType.Type.(type) {
		case *types.Named:
			if _, ok := t.Underlying().(*types.Struct); ok {
				structType = t
			}
		case *types.Pointer:
			if named, ok := t.Elem().(*types.Named); ok {
				if _, ok := named.Underlying().(*types.Struct); ok {
					structType = named
				}
			}
		}
	}

	if structType == nil {
		return
	}

	// Analyze the method body for field usage, assignments, and nil checks
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// Check for field access (e.g., o.field, self.field)
			if ident, ok := x.X.(*ast.Ident); ok {
				// Check if this is accessing the receiver
				if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
					if _, exists := fieldUsage[structType]; exists {
						fieldUsage[structType][x.Sel.Name] = true
					}
				}
			}
		case *ast.CallExpr:
			// Check for method calls that might indirectly use fields
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
						// This is a method call on the receiver, check if it uses fields indirectly
						p.analyzeMethodCallForFieldUsage(pass, sel.Sel.Name, structType, fieldUsage)
					}
				}
			}
		case *ast.AssignStmt:
			// Check for field assignments (e.g., o.field = value)
			for _, lhs := range x.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
							if _, exists := fieldAssignments[structType]; exists {
								fieldAssignments[structType][sel.Sel.Name] = true
							}
						}
					}
				}
			}
		case *ast.IfStmt:
			// Check for nil checks (e.g., if o.field == nil)
			p.analyzeNilCheck(x, recv, structType, nilSafeFields)
		}
		return true
	})
}

func (p *StructInitPlugin) analyzeMethodCallForFieldUsage(pass *analysis.Pass, methodName string, structType *types.Named, fieldUsage map[*types.Named]map[string]bool) {
	// Find the method definition and analyze it for field usage
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				// Check if this is a method of our struct type
				if fn.Recv != nil && len(fn.Recv.List) > 0 && fn.Name.Name == methodName {
					recv := fn.Recv.List[0]
					var methodStructType *types.Named

					if recvType := pass.TypesInfo.Types[recv.Type]; recvType.Type != nil {
						switch t := recvType.Type.(type) {
						case *types.Named:
							if _, ok := t.Underlying().(*types.Struct); ok {
								methodStructType = t
							}
						case *types.Pointer:
							if named, ok := t.Elem().(*types.Named); ok {
								if _, ok := named.Underlying().(*types.Struct); ok {
									methodStructType = named
								}
							}
						}
					}

					// If this method belongs to our struct, analyze its body for field usage
					if methodStructType == structType && fn.Body != nil {
						ast.Inspect(fn.Body, func(node ast.Node) bool {
							if sel, ok := node.(*ast.SelectorExpr); ok {
								if ident, ok := sel.X.(*ast.Ident); ok {
									if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
										// Mark this field as used
										if _, exists := fieldUsage[structType]; exists {
											fieldUsage[structType][sel.Sel.Name] = true
										}
									}
								}
							}
							return true
						})
					}
				}
			}
			return true
		})
	}
}

func (p *StructInitPlugin) analyzeNilCheck(ifStmt *ast.IfStmt, recv *ast.Field, structType *types.Named, nilSafeFields map[*types.Named]map[string]bool) {
	if ifStmt.Cond == nil {
		return
	}

	// Look for binary expressions like "field == nil" or "field != nil"
	if binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr); ok {
		var fieldName string
		var isNilCheck bool

		// Check for "o.field == nil" pattern
		if binExpr.Op == token.EQL || binExpr.Op == token.NEQ {
			if sel, ok := binExpr.X.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
						if nilIdent, ok := binExpr.Y.(*ast.Ident); ok && nilIdent.Name == "nil" {
							fieldName = sel.Sel.Name
							isNilCheck = true
						}
					}
				}
			}
			// Check for "nil == o.field" pattern
			if sel, ok := binExpr.Y.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if len(recv.Names) > 0 && ident.Name == recv.Names[0].Name {
						if nilIdent, ok := binExpr.X.(*ast.Ident); ok && nilIdent.Name == "nil" {
							fieldName = sel.Sel.Name
							isNilCheck = true
						}
					}
				}
			}
		}

		if isNilCheck && fieldName != "" {
			if _, exists := nilSafeFields[structType]; exists {
				nilSafeFields[structType][fieldName] = true
			}
		}
	}
}

func (p *StructInitPlugin) checkConstructor(pass *analysis.Pass, fn *ast.FuncDecl, structTypes map[*types.Named]map[string]*types.Var, fieldUsage map[*types.Named]map[string]bool, fieldAssignments map[*types.Named]map[string]bool, nilSafeFields map[*types.Named]map[string]bool) {
	if fn.Body == nil {
		return
	}

	// Find struct literal assignments in return statements
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			for _, result := range ret.Results {
				p.checkReturnExpression(pass, result, structTypes, fieldUsage, fieldAssignments, nilSafeFields)
			}
		}
		return true
	})
}

func (p *StructInitPlugin) checkReturnExpression(pass *analysis.Pass, expr ast.Expr, structTypes map[*types.Named]map[string]*types.Var, fieldUsage map[*types.Named]map[string]bool, fieldAssignments map[*types.Named]map[string]bool, nilSafeFields map[*types.Named]map[string]bool) {
	switch x := expr.(type) {
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			if comp, ok := x.X.(*ast.CompositeLit); ok {
				p.checkCompositeLit(pass, comp, structTypes, fieldUsage, fieldAssignments, nilSafeFields)
			}
		}
	case *ast.CompositeLit:
		p.checkCompositeLit(pass, x, structTypes, fieldUsage, fieldAssignments, nilSafeFields)
	}
}

func (p *StructInitPlugin) checkCompositeLit(pass *analysis.Pass, comp *ast.CompositeLit, structTypes map[*types.Named]map[string]*types.Var, fieldUsage map[*types.Named]map[string]bool, fieldAssignments map[*types.Named]map[string]bool, nilSafeFields map[*types.Named]map[string]bool) {
	// Get the type of this composite literal
	if compType := pass.TypesInfo.Types[comp]; compType.Type != nil {
		var structType *types.Named

		// Handle both direct struct type and pointer to struct
		switch t := compType.Type.(type) {
		case *types.Named:
			if _, ok := t.Underlying().(*types.Struct); ok {
				structType = t
			}
		case *types.Pointer:
			if named, ok := t.Elem().(*types.Named); ok {
				if _, ok := named.Underlying().(*types.Struct); ok {
					structType = named
				}
			}
		}

		if structType == nil {
			return
		}

		fields, exists := structTypes[structType]
		if !exists {
			return
		}

		// Collect initialized fields
		initializedFields := make(map[string]bool)

		// Handle both named field initialization and positional initialization
		if len(comp.Elts) > 0 {
			// Check if this is positional initialization (no key-value pairs)
			isPositional := true
			for _, elt := range comp.Elts {
				if _, ok := elt.(*ast.KeyValueExpr); ok {
					isPositional = false
					break
				}
			}

			if isPositional {
				// For positional initialization, mark fields as initialized based on position
				if structDef, ok := structType.Underlying().(*types.Struct); ok {
					for i := range comp.Elts {
						if i < structDef.NumFields() {
							field := structDef.Field(i)
							if !field.Embedded() {
								initializedFields[field.Name()] = true
							}
						}
					}
				}
			} else {
				// Handle named field initialization
				for _, elt := range comp.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if ident, ok := kv.Key.(*ast.Ident); ok {
							initializedFields[ident.Name] = true
						}
					}
				}
			}
		}

		// Check for fields that are used but not initialized or assigned
		for fieldName, field := range fields {
			isUsed := fieldUsage[structType][fieldName]
			isAssigned := fieldAssignments[structType][fieldName]
			isInitialized := initializedFields[fieldName]
			isNilSafe := nilSafeFields[structType][fieldName]

			// Only flag fields that are:
			// 1. Used in methods AND
			// 2. Not initialized in constructor AND
			// 3. Not dynamically assigned later AND
			// 4. Not nil-safe (checked for nil before use) AND
			// 5. Should be initialized (based on type)
			if isUsed && !isInitialized && !isAssigned && !isNilSafe && p.shouldBeInitialized(field) {
				pass.Report(analysis.Diagnostic{
					Pos:     comp.Pos(),
					Message: "struct field '" + fieldName + "' of type '" + field.Type().String() + "' is used but not initialized in constructor, not assigned dynamically, and not nil-safe",
				})
			}
		}
	}
}

func (p *StructInitPlugin) shouldBeInitialized(field *types.Var) bool {
	fieldType := field.Type().String()

	// Skip primitive types that have sensible zero values
	skipTypes := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"bool", "string",
		"sync.Mutex", "sync.RWMutex", "sync.WaitGroup",
		"sync/atomic.Bool", "sync/atomic.Int32", "sync/atomic.Int64",
		"sync/atomic.Uint32", "sync/atomic.Uint64",
		"embed.FS",
	}

	for _, skip := range skipTypes {
		if strings.Contains(fieldType, skip) {
			return false
		}
	}

	// Check if it's a context.Context (commonly left uninitialized)
	if strings.Contains(fieldType, "context.Context") {
		return false
	}

	// Interface types, pointers, and slices/maps should typically be initialized
	if strings.Contains(fieldType, "interface") ||
		strings.HasPrefix(fieldType, "*") ||
		strings.HasPrefix(fieldType, "[]") ||
		strings.HasPrefix(fieldType, "map[") {
		return true
	}

	// Named types from external packages should typically be initialized
	if strings.Contains(fieldType, ".") {
		return true
	}

	return false
}
