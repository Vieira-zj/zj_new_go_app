package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"demo.hello/utils"
)

// FuncInfo .
type FuncInfo struct {
	Path      string `json:"pkg_path"`
	Name      string `json:"name"`
	StartLine int    `json:"start_line"`
	StartCol  int    `json:"start_col"`
	EndLine   int    `json:"end_line"`
	EndCol    int    `json:"end_col"`
	StmtCount int    `json:"stmt_count"`
	Source    string `json:"source"`
}

//
// Inspects a func of go file.
//

// GetFuncInfo .
func GetFuncInfo(filePath string, src []byte, funcName string) (*FuncInfo, error) {
	fset := token.NewFileSet()
	root, err := getASTRoot(fset, filePath, src)
	if err != nil {
		return nil, err
	}

	rPath, err := getRelativePath(filePath, root.Name.Name)
	if err != nil {
		return nil, err
	}

	var funcInfo *FuncInfo
	ast.Inspect(root, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == funcName {
				start := fset.Position(fn.Pos())
				end := fset.Position(fn.End())
				funcInfo = &FuncInfo{
					Path:      rPath,
					Name:      fn.Name.Name,
					StartLine: start.Line,
					StartCol:  start.Column,
					EndLine:   end.Line,
					EndCol:    end.Column,
					StmtCount: len(fn.Body.List),
				}
				if err := addFuncInfoRecv(fset, fn, funcInfo); err != nil {
					log.Println(err.Error())
				}
				if err := addFuncInfoSource(fset, fn, funcInfo); err != nil {
					log.Println(err.Error())
				}
				return false
			}
		}
		return true
	})

	return funcInfo, nil
}

func getASTRoot(fset *token.FileSet, filePath string, src []byte) (*ast.File, error) {
	var (
		root *ast.File
		err  error
	)
	if len(filePath) > 0 {
		root, err = parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	} else if len(src) > 0 {
		root, err = parser.ParseFile(fset, "", src, parser.ParseComments)
	} else {
		return nil, fmt.Errorf("Invalid param [filePath] or [src]")
	}
	if err != nil {
		return nil, err
	}

	return root, nil
}

func addFuncInfoRecv(fset *token.FileSet, fn *ast.FuncDecl, funcInfo *FuncInfo) error {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		recvType, err := getASTFieldType(fset, recv)
		if err != nil {
			return fmt.Errorf("get rev field [%s] type error: %v", recv.Names[0], err)
		}
		funcInfo.Name = fmt.Sprintf("(%s %s) %s", recv.Names[0], recvType, funcInfo.Name)
	}
	return nil
}

func addFuncInfoSource(fset *token.FileSet, fn *ast.FuncDecl, funcInfo *FuncInfo) error {
	source, err := getFuncBodySource(fset, fn.Body)
	if err != nil {
		return fmt.Errorf("get func [%s] body source error: %v", fn.Name.Name, err)
	}
	funcInfo.Source = source
	return nil
}

func getFuncBodySource(fset *token.FileSet, body *ast.BlockStmt) (string, error) {
	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, body); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getASTFieldType(fset *token.FileSet, field *ast.Field) (string, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, field.Type); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type specFuncVisit struct {
	fset     *token.FileSet
	found    bool
	path     string
	funcInfo *FuncInfo
}

// Visit implements ast.Visitor interface.
func (v *specFuncVisit) Visit(n ast.Node) ast.Visitor {
	if n == nil || v.found {
		return v
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		if fn.Name.Name == v.funcInfo.Name {
			v.found = true
			start := v.fset.Position(fn.Pos())
			end := v.fset.Position(fn.End())
			funcInfo := &FuncInfo{
				Path:      v.path,
				Name:      fn.Name.Name,
				StartLine: start.Line,
				StartCol:  start.Column,
				EndLine:   end.Line,
				EndCol:    end.Column,
				StmtCount: len(fn.Body.List),
			}
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				recv := fn.Recv.List[0].Names[0].Name
				funcInfo.Name = fmt.Sprintf("(%s)%s", recv, funcInfo.Name)
			}
			v.funcInfo = funcInfo
		}
	}
	return v
}

// GetFuncInfoDeprecated returns func info: start line:col, end line:col, and total statements.
func GetFuncInfoDeprecated(filePath, funcName string) (*FuncInfo, error) {
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	rPath, err := getRelativePath(filePath, root.Name.Name)
	if err != nil {
		return nil, err
	}

	visit := &specFuncVisit{
		fset: fset,
		path: rPath,
		funcInfo: &FuncInfo{
			Name: funcName,
		},
	}
	ast.Walk(visit, root)
	return visit.funcInfo, nil
}

// GetFuncSrc returns func source between start line:col and end line:col.
func GetFuncSrc(src []byte, info *FuncInfo) []byte {
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
	return res
}

//
// Walks each func of go file.
//

type funcVisit struct {
	fset      *token.FileSet
	path      string
	funcInfos []*FuncInfo
}

// Visit implements ast.Visitor interface.
func (v *funcVisit) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}

	if fn, ok := n.(*ast.FuncDecl); ok {
		funcInfo := v.parseFuncDecl(fn)
		v.funcInfos = append(v.funcInfos, funcInfo)
	} else if fn, ok := n.(*ast.FuncLit); ok {
		funcInfo, err := v.parseFuncLit(fn)
		if err != nil {
			log.Println(err)
			return v
		}
		log.Println("ignore anonymous func:", prettySprintFuncInfo(funcInfo))
	}
	return v
}

func (v *funcVisit) parseFuncDecl(fn *ast.FuncDecl) *FuncInfo {
	start := v.fset.Position(fn.Pos())
	end := v.fset.Position(fn.End())
	funcInfo := &FuncInfo{
		Path:      v.path,
		Name:      fn.Name.Name,
		StartLine: start.Line,
		StartCol:  start.Column,
		EndLine:   end.Line,
		EndCol:    end.Column,
		StmtCount: len(fn.Body.List),
	}
	if err := addFuncInfoRecv(v.fset, fn, funcInfo); err != nil {
		log.Println(err.Error())
	}
	if err := addFuncInfoSource(v.fset, fn, funcInfo); err != nil {
		log.Println(err.Error())
	}
	return funcInfo
}

func (v *funcVisit) parseFuncLit(fn *ast.FuncLit) (*FuncInfo, error) {
	randID, err := utils.GetRandString(8)
	if err != nil {
		return nil, err
	}

	start := v.fset.Position(fn.Pos())
	end := v.fset.Position(fn.End())
	funcInfo := &FuncInfo{
		Path:      v.path,
		Name:      "anonymous_" + randID,
		StartLine: start.Line,
		StartCol:  start.Column,
		EndLine:   end.Line,
		EndCol:    end.Column,
		StmtCount: len(fn.Body.List),
	}
	return funcInfo, nil
}

func (v *funcVisit) parseFuncLitFromAssignStmt(assign *ast.AssignStmt) *FuncInfo {
	if fn, ok := assign.Rhs[0].(*ast.FuncLit); ok {
		name := "anonymous_"
		if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
			name = name + ident.Name
		}

		start := v.fset.Position(fn.Pos())
		end := v.fset.Position(fn.End())
		funcInfo := &FuncInfo{
			Path:      v.path,
			Name:      name,
			StartLine: start.Line,
			StartCol:  start.Column,
			EndLine:   end.Line,
			EndCol:    end.Column,
			StmtCount: len(fn.Body.List),
		}
		return funcInfo
	}
	return nil
}

// GetFuncInfos returns all funcs info of go file.
func GetFuncInfos(filePath string, src []byte) ([]*FuncInfo, error) {
	fset := token.NewFileSet()
	root, err := getASTRoot(fset, filePath, src)
	if err != nil {
		return nil, err
	}

	rPath, err := getRelativePath(filePath, root.Name.Name)
	if err != nil {
		return nil, err
	}

	visit := &funcVisit{
		fset: fset,
		path: rPath,
	}
	ast.Walk(visit, root)
	return visit.funcInfos, nil
}

//
// Format .go file.
//

// formatGoFile filters out empty and comment lines for .go files.
func formatGoFile(src, dst string) error {
	comments, err := getCommentsInGoFile(src)
	if err != nil {
		return err
	}

	lines, err := utils.ReadLinesFile(src)
	if err != nil {
		return err
	}

	outLines := make([]string, 0, len(lines))
	for _, line := range lines {
		newLine := line
		for len(comments) > 0 {
			comment := comments[0]
			if strings.Index(newLine, comment) == -1 {
				break
			}
			newLine = strings.Replace(newLine, comment, "", 1)
			newLine = strings.Trim(newLine, " ")
			newLine = strings.Trim(newLine, "\t")
			newLine = strings.Trim(newLine, "\n")
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

func getCommentsInGoFile(path string) ([]string, error) {
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

func runGoFmt(filePath string) error {
	_, err := utils.RunShellCmd("go", "fmt", filePath)
	return err
}

func getRelativePath(path, pkgName string) (string, error) {
	pkg, err := getGoPackage(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	if pkgName != "main" && !strings.HasSuffix(pkg, pkgName) {
		return "", fmt.Errorf("Package name inconsistent")
	}
	return filepath.Join(pkg, filepath.Base(path)), nil
}

func getGoPackage(dirPath string) (string, error) {
	sh := utils.GetShellPath()
	res, err := utils.RunShellCmd(sh, "-c", fmt.Sprintf("cd %s && go list .", dirPath))
	if err != nil {
		return "", err
	}

	res = strings.Trim(res, "\n")
	res = strings.Trim(res, " ")
	return res, nil
}

func deleteEmptyLinesInText(src []byte) string {
	lines := strings.Split(string(src), "\n")
	outLines := make([]string, 0, len(lines))
	for _, line := range lines {
		newLine := strings.Trim(line, " ")
		newLine = strings.Trim(newLine, "\t")
		newLine = strings.Trim(newLine, "\n")
		if len(newLine) > 0 {
			outLines = append(outLines, line)
		}
	}

	return strings.Join(outLines, "\n")
}
