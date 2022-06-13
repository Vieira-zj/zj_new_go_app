package funcdiff

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"demo.hello/utils"
)

// FuncInfo .
type FuncInfo struct {
	Package             string
	FuncName            string
	StartLine, StartCol int
	EndLine, EndCol     int
	StmtCount           int
}

//
// Walks specified func of go file.
//

// FuncVisit .
type FuncVisit struct {
	fset  *token.FileSet
	found bool
	pkg   string
	Info  *FuncInfo
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
				Package:   v.pkg,
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
func GetFuncInfo(filePath, funcName string) (*FuncInfo, error) {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	visit := &FuncVisit{
		fset: fset,
		pkg:  root.Name.Name,
		Info: &FuncInfo{
			FuncName: funcName,
		},
	}
	ast.Walk(visit, root)
	return visit.Info, nil
}

// GetFuncSrc returns func source between start line:col and end line:col.
func GetFuncSrc(src []byte, info *FuncInfo) string {
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
	return string(res)
}

//
// Walks all funcs of go file.
//

// AllFuncsVisit .
type AllFuncsVisit struct {
	fset  *token.FileSet
	pkg   string
	Infos []*FuncInfo
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
			Package:   v.pkg,
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

// GetFuncInfos returns all funcs info of go file.
func GetFuncInfos(path string) ([]*FuncInfo, error) {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	visit := &AllFuncsVisit{
		fset: fset,
		pkg:  root.Name.Name,
	}
	ast.Walk(visit, root)
	return visit.Infos, nil
}

//
// Format .go files.
//

// FormatGoFile filters out empty and comment lines for .go files.
func FormatGoFile(src, dst string) error {
	comments, err := GetComments(src)
	if err != nil {
		return err
	}

	lines, err := utils.ReadLinesFile(src)
	if err != nil {
		return err
	}

	var outLines []string
	for _, line := range lines {
		newLine := line
		for len(comments) > 0 {
			comment := comments[0]
			if strings.Index(newLine, comment) == -1 {
				break
			}
			newLine = strings.Replace(newLine, comment, "", 1)
			newLine = strings.Trim(newLine, " ")
			comments = comments[1:]
		}
		if len(newLine) > 0 {
			outLines = append(outLines, newLine)
		}
	}

	if err := os.Remove(dst); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	out := []byte(strings.Join(outLines, "\n"))
	if err := os.WriteFile(dst, out, 0644); err != nil {
		return err
	}

	return runGoFmt(dst)
}

// GetComments .
func GetComments(path string) ([]string, error) {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	comments := make([]string, 0, 16)
	for _, group := range root.Comments {
		for _, comment := range group.List {
			comments = append(comments, comment.Text)
		}
	}
	return comments, nil
}

func runGoFmt(path string) error {
	_, err := utils.RunShellCmd("gofmt", "-w", path)
	return err
}
