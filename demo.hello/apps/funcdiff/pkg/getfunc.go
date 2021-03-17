package funcdiff

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

/*
A specified function of go file.
*/

// FuncInfo func info data.
type FuncInfo struct {
	FuncName            string
	StartLine, StartCol int
	EndLine, EndCol     int
	StmtCount           int
}

// FuncVisit get specified func info.
type FuncVisit struct {
	fset    *token.FileSet
	found   bool
	Package string
	Info    *FuncInfo
}

// Visit implements ast.Visitor interface.
func (v *FuncVisit) Visit(n ast.Node) ast.Visitor {
	if n == nil || v.found {
		return v
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		if fn.Name.Name == v.Info.FuncName {
			v.found = true
			start := v.fset.Position(fn.Pos())
			end := v.fset.Position(fn.End())
			v.Info = &FuncInfo{
				StartLine: start.Line,
				StartCol:  start.Column,
				EndLine:   end.Line,
				EndCol:    end.Column,
				StmtCount: len(fn.Body.List),
			}
		}
	}
	return v
}

// GetFuncInfo returns func info: start line:col, end line:col, and total statements.
func GetFuncInfo(path, funcName string) (*FuncInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	v := &FuncVisit{
		fset:    fset,
		Package: f.Name.Name,
		Info: &FuncInfo{
			FuncName: funcName,
		},
	}
	ast.Walk(v, f)
	return v.Info, nil
}

// GetFuncSrc returns func source between start line:col and end line:col.
func GetFuncSrc(path string, info *FuncInfo) (string, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	var res []byte
	line := 1
	col := 1
	for charIdx := 0; charIdx < len(src); charIdx++ {
		if (line == info.EndLine && col > info.EndCol) || (line > info.EndLine) {
			break
		}
		if (line == info.StartLine && col >= info.StartCol) ||
			(line > info.StartLine && line < info.EndLine) ||
			(line == info.EndLine && col < info.EndCol) {
			res = append(res, src[charIdx])
		}

		if src[charIdx] == '\n' {
			line++
			col = 0
		}
		col++
	}
	return string(res), nil
}

/*
All functions of go file.
*/

// AllFuncsVisit walks all funcs info of go file.
type AllFuncsVisit struct {
	fset    *token.FileSet
	Package string
	Infos   []*FuncInfo
}

// Visit implements ast.Visitor interface.
func (v *AllFuncsVisit) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		start := v.fset.Position(fn.Pos())
		end := v.fset.Position(fn.End())
		info := &FuncInfo{
			FuncName:  fn.Name.Name,
			StartLine: start.Line,
			StartCol:  start.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
			StmtCount: len(fn.Body.List),
		}
		v.Infos = append(v.Infos, info)
	}
	return v
}

// GetFileAllFuncInfos returns all funcs info of go file.
func GetFileAllFuncInfos(path string) ([]*FuncInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	v := &AllFuncsVisit{
		fset:    fset,
		Package: f.Name.Name,
	}
	ast.Walk(v, f)
	return v.Infos, nil
}
