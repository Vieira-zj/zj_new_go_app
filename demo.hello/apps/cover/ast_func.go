package cover

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// FuncExtent describes a function's extent in the source by file and position.
type FuncExtent struct {
	name      string
	startLine int
	startCol  int
	endLine   int
	endCol    int
}

// FuncVisitor implements the visitor that builds the function position list for a file.
type FuncVisitor struct {
	fset    *token.FileSet
	name    string // Name of file
	astFile *ast.File
	funcs   []*FuncExtent
}

// Visit implements the ast.Visitor interface.
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		start := v.fset.Position(n.Pos())
		end := v.fset.Position(n.End())
		fe := &FuncExtent{
			name:      n.Name.Name,
			startLine: start.Line,
			startCol:  start.Column,
			endLine:   end.Line,
			endCol:    end.Column,
		}
		v.funcs = append(v.funcs, fe)
	}

	return v
}

// findFuncs parses go file and returns a slice of FuncExtent descriptors by ast visitor.
func findFuncs(filePath string) ([]*FuncExtent, error) {
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	visitor := &FuncVisitor{
		fset:    fset,
		name:    filePath,
		astFile: parsedFile,
	}
	ast.Walk(visitor, visitor.astFile)
	return visitor.funcs, nil
}

// getCoverage returns coverage data by compare fe and profile blocks.
// profile: .cov -> []*Profile -> Profile -> .go file -> coverage blocks
// fe: .cov -> .go file -> ast -> []*FuncExtent -> FuncExtent
// fe -> matched coverage blocks
func (fe *FuncExtent) getCoverage(profile *Profile) (int64, int64) {
	var covered, total int64

	for _, b := range profile.Blocks {
		// past the end of the function
		if b.StartLine > fe.endLine || (b.StartLine == fe.endLine && b.StartCol >= fe.endCol) {
			break
		}
		// before the beginning of the function
		if b.EndLine < fe.startLine || (b.EndLine == fe.startLine && b.EndCol <= fe.startCol) {
			continue
		}
		// print matched pair of block and func
		fmt.Printf("%s:%d\t%s\n", profile.FileName, b.StartLine, fe.name)
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}

	// Avoid zero denominator
	if total == 0 {
		total = 1
	}
	return covered, total
}
