package pkg

import (
	"fmt"
	"go/build"
	"log"
	"time"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

/*
函数调用结构体定义
*/

// FuncDesc 函数定义
type FuncDesc struct {
	File    string // 文件路径
	Package string // package名
	Name    string // 函数名，格式为Package.Func
}

// CallerRelation 描述一个函数调用N个函数的一对多关系
type CallerRelation struct {
	Caller  FuncDesc
	Callees []FuncDesc
}

// ReverseCallRelation 描述关键函数的一条反向调用关系
type ReverseCallRelation struct {
	Callees []FuncDesc
	CanFix  bool // 该调用关系能反向找到gin.Context即可以自动修复
}

// Fixed 关键函数定义
type Fixed struct {
	FuncDesc
	RelationsTree *MWTNode // 反向调用关系，可能有多条调用链到达关键函数
	RelationList  []ReverseCallRelation
	CanFix        bool // 能反向找到gin.Context即可以自动修复
}

/*
函数分析
go/loader: Package loader loads a complete Go program from source code,
parsing and type-checking the initial packages plus their transitive closure of dependencies.

go/ssa: Package ssa defines a representation of the elements of Go programs (packages, types, functions, variables and constants)
using a static single-assignment (SSA) form intermediate representation (IR) for the bodies of functions.

go/pointer: Package pointer implements Andersen's analysis, an inclusion-based pointer analysis algorithm.
（指针分析是一类特殊的数据流问题，它是其它静态程序分析的基础。算法最终建立各节点间的指向关系。）
*/

// Analysis includes analysis data and results.
type Analysis struct {
	prog   *ssa.Program
	conf   loader.Config
	pkgs   []*ssa.Package
	mains  []*ssa.Package
	result *pointer.Result
}

// RunAnalysis runs go package analysis.
func RunAnalysis(buildCtx *build.Context, tests bool, args []string) (*Analysis, error) {
	t0 := time.Now()
	conf := loader.Config{Build: buildCtx}
	_, err := conf.FromArgs(args, tests)
	if err != nil {
		return nil, fmt.Errorf("invalid args: %v", args)
	}

	load, err := conf.Load()
	if err != nil {
		return nil, fmt.Errorf("failed conf load: %v", err)
	}
	log.Printf("loading.. %d imported (%d created) took: %v", len(load.Imported), len(load.Created), time.Since(t0))

	t0 = time.Now()
	prog := ssautil.CreateProgram(load, 0)
	prog.Build()
	pkgs := prog.AllPackages()

	var mains []*ssa.Package
	if tests {
		for _, pkg := range pkgs {
			if main := prog.CreateTestMainPackage(pkg); main != nil {
				mains = append(mains, main)
			}
		}
		if mains == nil {
			log.Fatalln("no tests")
		}
	} else {
		mains = append(mains, ssautil.MainPackages(pkgs)...)
		if len(mains) == 0 {
			log.Printf("no main packages")
		}
	}
	log.Printf("building.. %d packages (%d main) took: %v", len(pkgs), len(mains), time.Since(t0))

	t0 = time.Now()
	ptrcfg := &pointer.Config{
		Mains:          mains,
		BuildCallGraph: true,
	}
	result, err := pointer.Analyze(ptrcfg)
	if err != nil {
		log.Fatalln("analyze failed:", err)
	}
	log.Printf("analysis took: %v", time.Since(t0))

	return &Analysis{
		prog:   prog,
		conf:   conf,
		pkgs:   pkgs,
		mains:  mains,
		result: result,
	}, nil
}

// RenderOpts go package analysis render options.
type RenderOpts struct {
	Nointer bool
	Nostd   bool
	Focus   string
	Ignore  []string
	Include []string
}

// Render returns analysis callmap results.
func (a *Analysis) Render(project string, opts *RenderOpts) (map[string]CallerRelation, error) {
	log.Printf("no std packages: %v", opts.Nostd)
	log.Printf("%d include prefixes: %v", len(opts.Include), opts.Include)
	log.Printf("%d ignore prefixes: %v", len(opts.Ignore), opts.Ignore)
	return printOutput(a.prog, a.result.CallGraph, opts)
}
