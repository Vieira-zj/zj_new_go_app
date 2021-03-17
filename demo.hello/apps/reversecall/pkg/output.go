package pkg

import (
	"go/build"
	"log"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

func isSynthetic(edge *callgraph.Edge) bool {
	return edge.Caller.Func.Pkg == nil || edge.Callee.Func.Synthetic != ""
}

func inStd(node *callgraph.Node) bool {
	pkg, _ := build.Import(node.Func.Pkg.Pkg.Path(), "", 0)
	return pkg.Goroot
}

func printOutput(prog *ssa.Program, cg *callgraph.Graph, opts *RenderOpts) (map[string]CallerRelation, error) {
	callMap := make(map[string]CallerRelation, 0)
	cg.DeleteSyntheticNodes()

	var isFocused = func(edge *callgraph.Edge) bool {
		caller := edge.Caller
		callee := edge.Callee
		if caller.Func.Pkg.Pkg.Path() == opts.Focus ||
			callee.Func.Pkg.Pkg.Path() == opts.Focus {
			return true
		}
		fromFocused := false
		toFocused := false
		for _, e := range caller.In {
			if !isSynthetic(e) && e.Caller.Func.Pkg.Pkg.Path() == opts.Focus {
				fromFocused = true
				break
			}
		}
		for _, e := range callee.Out {
			if !isSynthetic(e) && e.Callee.Func.Pkg.Pkg.Path() == opts.Focus {
				toFocused = true
				break
			}
		}
		if fromFocused && toFocused {
			log.Printf("edge semi-focus: %s", edge)
			return true
		}
		return false
	}

	var isInter = func(edge *callgraph.Edge) bool {
		//caller := edge.Caller
		callee := edge.Callee
		if callee.Func.Object() != nil && !callee.Func.Object().Exported() {
			return true
		}
		return false
	}

	var inIncludes = func(node *callgraph.Node) bool {
		pkgPath := node.Func.Pkg.Pkg.Path()
		for _, p := range opts.Include {
			if strings.HasPrefix(pkgPath, p) {
				return true
			}
		}
		return false
	}

	var inIgnores = func(node *callgraph.Node) bool {
		pkgPath := node.Func.Pkg.Pkg.Path()
		for _, p := range opts.Ignore {
			if strings.HasPrefix(pkgPath, p) {
				return true
			}
		}
		return false
	}

	count := 0
	var onGraphVisitEdges = func(edge *callgraph.Edge) error {
		count++

		caller := edge.Caller
		callee := edge.Callee

		callerPos := prog.Fset.Position(caller.Func.Pos())
		callerFile := callerPos.Filename

		calleePos := prog.Fset.Position(callee.Func.Pos())
		calleeFile := calleePos.Filename

		if strings.Contains(callerFile, "vendor") || strings.Contains(calleeFile, "vendor") {
			return nil
		}
		// omit synthetic calls
		if isSynthetic(edge) {
			return nil
		}
		// omit std
		if opts.Nostd && (inStd(caller) || inStd(callee)) {
			return nil
		}
		// omit inter
		if opts.Nointer && isInter(edge) {
			return nil
		}

		// focus specific pkg
		if len(opts.Focus) > 0 && !isFocused(edge) {
			return nil
		}

		// include path prefixes
		if len(opts.Include) > 0 && (!inIncludes(caller) || !inIncludes(callee)) {
			// log.Printf("NOT in include: %s -> %s", caller, callee)
			return nil
		}
		// ignore path prefixes
		if len(opts.Ignore) > 0 && (inIgnores(caller) || inIgnores(callee)) {
			// log.Printf("IS ignored: %s -> %s", caller, callee)
			return nil
		}

		// var buf bytes.Buffer
		// data, _ := json.MarshalIndent(caller.Func, "", " ")
		// log.Printf("call node: %s -> %s\n %v", caller, callee, string(data))
		// log.Printf("package: %s -> %s (%s -> %s)", caller.Func.Pkg.Pkg.Name(), callee.Func.Pkg.Pkg.Name(), caller.Func.Object().Name(), caller.Func.Name(), callee.Func.Name())
		log.Printf("call node: %s -> %s", caller.String(), callee.String())

		callerPkg := caller.Func.Pkg.Pkg.Name()
		calleePkg := callee.Func.Pkg.Pkg.Name()
		callerName := strings.Split(caller.String(), "/")[len(strings.Split(caller.String(), "/"))-1]
		calleeName := strings.Split(callee.String(), "/")[len(strings.Split(callee.String(), "/"))-1]

		// 针对 go func(){} 的情况，处理 $ (比如 Test3c$1)
		callerName = strings.Split(callerName, "$")[0]
		calleeName = strings.Split(calleeName, "$")[0]

		// 防止递归
		if callerName == calleeName {
			log.Printf("recursion call:%s", callerName)
			return nil
		}

		// 注意类的方法, 表现形式不一样: (demo.hello/apps/calltrace/test/example.XYZ).print
		// callerList和calleeList的第一个元素是package, 第二个元素是function (包括类的function): ["example", "XYZ@print"]
		if strings.Contains(callerName, ").") {
			callerName = strings.Replace(callerName, ").", "@", -1)
		}
		callerList := strings.Split(callerName, ".")
		if strings.Contains(calleeName, ").") {
			calleeName = strings.Replace(calleeName, ").", "@", -1)
		}
		calleeList := strings.Split(calleeName, ".")

		if v, ok := callMap[callerName]; ok {
			for _, c := range v.Callees {
				if c.Package == calleeList[0] && c.Name == calleeList[1] {
					log.Printf("duplicated call node: %s -> %s", caller, callee)
					return nil
				}
			}
			list := append(v.Callees, FuncDesc{
				calleeFile,
				calleePkg,
				calleeList[1]})
			v.Callees = list
			callMap[callerName] = v
		} else {
			callMap[callerName] = CallerRelation{
				Caller: FuncDesc{
					callerFile,
					callerPkg,
					callerList[1]},
				Callees: []FuncDesc{{
					calleeFile,
					calleePkg,
					calleeList[1]}}}
		}

		return nil
	}

	// 深度优先遍历 callgraph.Graph 得到函数之间的两两调用关系
	if err := callgraph.GraphVisitEdges(cg, onGraphVisitEdges); err != nil {
		return nil, err
	}

	log.Printf("%d/%d edges", len(callMap), count)
	return callMap, nil
}
