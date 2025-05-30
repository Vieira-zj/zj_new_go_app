package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
)

//
// Ast Func Parser
//

// Fixed 关键函数定义
type Fixed struct {
	FuncDesc
}

// FuncDesc 函数定义
type FuncDesc struct {
	File    string // 文件路径
	Package string // package名
	Name    string // 函数名，格式为Package.Func
}

// GFset global fset
var GFset *token.FileSet

// GFixedFunc global fixed func
var GFixedFunc map[string]Fixed // key的格式为Package.Func

func stmtCase(stmt ast.Stmt, todo func(call *ast.CallExpr) bool) bool {
	// CallExpr 调用类型，类似于 "expr()"
	switch t := stmt.(type) {
	case *ast.ExprStmt:
		log.Printf("表达式语句%+v at line:%v", t, GFset.Position(t.Pos()))
		if call, ok := t.X.(*ast.CallExpr); ok {
			return todo(call)
		}
	case *ast.ReturnStmt:
		for i, p := range t.Results {
			log.Printf("return语句%d:%v at line:%v", i, p, GFset.Position(p.Pos()))
			if call, ok := p.(*ast.CallExpr); ok {
				return todo(call)
			}
		}
	// 函数体里的构造类型 9
	case *ast.AssignStmt:
		// Rhs 右表达式
		for _, p := range t.Rhs {
			switch t := p.(type) {
			// 构造类型 {}
			case *ast.CompositeLit:
				for i, p := range t.Elts {
					switch t := p.(type) {
					case *ast.KeyValueExpr:
						log.Printf("构造赋值语句%d:%+v at line:%v", i, t.Value, GFset.Position(p.Pos()))
						if call, ok := t.Value.(*ast.CallExpr); ok {
							return todo(call)
						}
					}
				}
			}
		}
	default:
		log.Printf("不匹配的类型:%T", stmt)
	}

	return false
}

// AllCallCase 查找逻辑
func AllCallCase(n ast.Node, todo func(call *ast.CallExpr) bool) (find bool) {
	// 函数体里的直接调用 0
	if fn, ok := n.(*ast.FuncDecl); ok {
		for i, p := range fn.Body.List {
			log.Printf("函数体表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
			find = stmtCase(p, todo) || find
		}
		log.Printf("func结束:%+v", fn.Name.Name)
	}

	// if语句里 1
	if ifstmt, ok := n.(*ast.IfStmt); ok {
		log.Printf("if语句开始:%T %+v", ifstmt, GFset.Position(ifstmt.If))
		if a, ok := ifstmt.Init.(*ast.AssignStmt); ok {
			for i, p := range a.Rhs {
				log.Printf("if语句赋值%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
				switch call := p.(type) {
				case *ast.CallExpr:
					c := todo(call)
					find = find || c
				}
			}
		}

		// if的花括号里面 2
		for i, p := range ifstmt.Body.List {
			log.Printf("if语句内部表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
			c := stmtCase(p, todo)
			find = find || c
		}

		// if的else里面 3
		if b, ok := ifstmt.Else.(*ast.BlockStmt); ok {
			for i, p := range b.List {
				log.Printf("if语句else表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
				c := stmtCase(p, todo)
				find = find || c
			}
		}

		log.Printf("if语句结束:%+v", GFset.Position(ifstmt.End()))
	}

	// 赋值语句 4
	if assign, ok := n.(*ast.AssignStmt); ok {
		log.Printf("赋值语句开始:%T %s", assign, GFset.Position(assign.Pos()))
		for i, p := range assign.Rhs {
			log.Printf("赋值表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
			switch t := p.(type) {
			case *ast.CallExpr:
				c := todo(t)
				find = find || c
			case *ast.CompositeLit:
				for i, p := range t.Elts {
					switch t := p.(type) {
					case *ast.KeyValueExpr:
						log.Printf("构造赋值%d:%+v at line:%v", i, t.Value, GFset.Position(p.Pos()))
						if call, ok := t.Value.(*ast.CallExpr); ok {
							c := todo(call)
							find = find || c
						}
					}
				}
			}
		}
	}

	if gostmt, ok := n.(*ast.GoStmt); ok {
		log.Printf("go语句开始:%T %s", gostmt.Call.Fun, GFset.Position(gostmt.Go))

		// go后面直接调用 5
		c := todo(gostmt.Call)
		find = find || c

		// go func里面的调用 6
		// FuncLit 函数定义
		if g, ok := gostmt.Call.Fun.(*ast.FuncLit); ok {
			for i, p := range g.Body.List {
				log.Printf("go语句表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
				c := stmtCase(p, todo)
				find = find || c
			}

			log.Printf("go语句结束:%+v", GFset.Position(gostmt.Go))
		}
	}

	if deferstmt, ok := n.(*ast.DeferStmt); ok {
		log.Printf("defer语句开始:%T %s", deferstmt.Call.Fun, GFset.Position(deferstmt.Defer))

		// defer后面直接调用 7
		c := todo(deferstmt.Call)
		find = find || c

		// defer func里面的调用 8
		if g, ok := deferstmt.Call.Fun.(*ast.FuncLit); ok {
			for i, p := range g.Body.List {
				log.Printf("defer语句内部表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
				c := stmtCase(p, todo)
				find = find || c
			}
		}

		log.Printf("defer语句结束:%+v", GFset.Position(deferstmt.Defer))
	}

	if fostmt, ok := n.(*ast.ForStmt); ok {
		// for语句对应 a 和 b
		log.Printf("for语句开始:%T %s", fostmt.Body, GFset.Position(fostmt.Pos()))
		for i, p := range fostmt.Body.List {
			log.Printf("for语句函数体表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
			c := stmtCase(p, todo)
			find = find || c
		}
	}

	if rangestmt, ok := n.(*ast.RangeStmt); ok {
		//range语句对应 c
		log.Printf("range语句开始:%T %s", rangestmt.Body, GFset.Position(rangestmt.Pos()))
		for i, p := range rangestmt.Body.List {
			log.Printf("range语句函数体表达式%d:%T at line:%v", i, p, GFset.Position(p.Pos()))
			c := stmtCase(p, todo)
			find = find || c
		}
	}
	return
}

// FindContext ast find context.
type FindContext struct {
	File      string
	Package   string
	LocalFunc *ast.FuncDecl
}

// Visit ast walk visit (ast.Visitor接口).
func (f *FindContext) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return f
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		log.Printf("函数 [%s.%s] 开始 at line:%v", f.Package, fn.Name.Name, GFset.Position(fn.Pos()))
		f.LocalFunc = fn
	} else {
		log.Printf("类型 %T at line:%v", n, GFset.Position(n.Pos()))
	}

	find := AllCallCase(n, f.FindCallFunc)
	if find {
		name := fmt.Sprintf("%s.%s", f.Package, f.LocalFunc.Name)
		GFixedFunc[name] = Fixed{FuncDesc: FuncDesc{f.File, f.Package, f.LocalFunc.Name.Name}}
	}
	return f
}

// FindCallFunc 查找 context.WithCancel 函数，并且入参为 nil
func (f *FindContext) FindCallFunc(call *ast.CallExpr) bool {
	if call == nil {
		return false
	}

	log.Printf("call func:%+v, %v", call.Fun, call.Args)

	// SelectorExpr 选择结构，类似于 "a.b" 的结构
	if callFunc, ok := call.Fun.(*ast.SelectorExpr); ok {
		if fmt.Sprint(callFunc.X) == "context" && fmt.Sprint(callFunc.Sel) == "WithCancel" {
			if len(call.Args) > 0 {
				// Ident 变量名
				if argu, ok := call.Args[0].(*ast.Ident); ok {
					log.Printf("argu type:%T, %s", argu.Name, argu.String())
					if argu.Name == "nil" {
						location := fmt.Sprint(GFset.Position(argu.NamePos))
						log.Printf("找到关键函数:%s.%s at line:%v", callFunc.X, callFunc.Sel, location)
						return true
					}
				}
			}
		}
	}
	return false
}

//
// Ast Generate Code
//

// GeneVisit .
type GeneVisit struct {
	fset    *token.FileSet
	Package string
}

// Visit .
func (v *GeneVisit) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		if fn.Name.Name == "main" {
			v.createSourceCtx(fn)
		} else {
			v.insertCtxIntoFuncParams(fn)
		}
	} else if stmt, ok := n.(*ast.ExprStmt); ok {
		if call, ok := stmt.X.(*ast.CallExpr); ok {
			v.insertCtxIntoCallArgs(call)
		}
	}
	return v
}

func (v *GeneVisit) insertCtxIntoCallArgs(call *ast.CallExpr) bool {
	if len(call.Args) > 0 {
		if arg, ok := call.Args[0].(*ast.Ident); ok {
			if arg.Name == "ctx" {
				log.Printf("call [%v] already have ctx in args.", call.Fun)
				return false
			}
			if arg.Name == "nil" {
				call.Args[0].(*ast.Ident).Name = "ctx"
				log.Printf("call [%v] replace nil to ctx.", call.Fun)
				return true
			}
		}
	}

	// src: ctx
	newArgs := make([]ast.Expr, 0, len(call.Args)+1)
	newArg := &ast.Ident{
		Name:    "ctx",
		Obj:     ast.NewObj(ast.Var, "ctx"),
		NamePos: call.Pos() + 1,
	}
	newArgs = append(newArgs, newArg)
	newArgs = append(newArgs, call.Args...)
	call.Args = newArgs

	log.Printf("call [%s] insert 1st arg ctx.", call.Fun)
	return true
}

func (v *GeneVisit) insertCtxIntoFuncParams(fn *ast.FuncDecl) bool {
	if len(fn.Type.Params.List) > 0 {
		firstParam := fn.Type.Params.List[0]
		if firstParam.Names[0].Name == "ctx" {
			log.Printf("func [%s] already have context in params", fn.Name)
			return false
		}
	}

	// src: ctx context.Context
	newParamName := &ast.Ident{
		Name:    "ctx",
		Obj:     ast.NewObj(ast.Var, "ctx"),
		NamePos: fn.Body.Pos() + 1,
	}
	newParamType := &ast.Ident{
		Name:    "context.Context",
		NamePos: newParamName.End() + 1,
	}

	newParams := make([]*ast.Field, 0, len(fn.Type.Params.List)+1)
	newParams = append(newParams, &ast.Field{
		Names: []*ast.Ident{newParamName},
		Type:  newParamType,
	})
	newParams = append(newParams, fn.Type.Params.List...)
	fn.Type.Params.List = newParams

	log.Printf("func [%s] insert 1st param ctx.", fn.Name.Name)
	return true
}

func (v *GeneVisit) createSourceCtx(fn *ast.FuncDecl) bool {
	for _, stmt := range fn.Body.List {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			for _, p := range assign.Lhs {
				if fmt.Sprint(p) == "ctx" {
					log.Printf("func [%s] already has create context.", fn.Name.Name)
					return false
				}
			}
		}
	}

	// src: ctx := context.Background()
	lhs := &ast.Ident{
		Name:    "ctx",
		Obj:     ast.NewObj(ast.Var, "ctx"),
		NamePos: fn.Body.Pos() + 1,
	}

	x := &ast.Ident{
		Name:    "context",
		Obj:     ast.NewObj(ast.Var, "context"),
		NamePos: fn.Body.Pos() + 1 + token.Pos(len("ctx := ")),
	}
	sel := &ast.Ident{
		Name:    "Background",
		Obj:     ast.NewObj(ast.Var, "Background"),
		NamePos: fn.Body.Pos() + 1 + token.Pos(len("ctx := context.")),
	}
	call := &ast.SelectorExpr{
		X:   x,
		Sel: sel,
	}
	rhs := &ast.CallExpr{
		Fun:    call,
		Args:   []ast.Expr{},
		Lparen: fn.Body.Pos() + token.Pos(len("ctx := context.Background(")+1),
		Rparen: fn.Body.Pos() + token.Pos(len("ctx := context.Background()")+1),
	}

	assign := &ast.AssignStmt{
		Lhs:    []ast.Expr{lhs},
		Rhs:    []ast.Expr{rhs},
		TokPos: lhs.Pos() + 1,
		Tok:    token.DEFINE,
	}

	newBody := make([]ast.Stmt, 0, len(fn.Body.List)+1)
	newBody = append(newBody, assign)
	newBody = append(newBody, fn.Body.List...)
	fn.Body.List = newBody

	log.Printf("func [%s] create ctx init stmt.", fn.Name.Name)
	return true
}

func formatAndGenerateGoFile(path string, src []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(src); err != nil {
		return err
	}

	// cmd: go fmt [path] && goimports -w [path]
	return nil
}
