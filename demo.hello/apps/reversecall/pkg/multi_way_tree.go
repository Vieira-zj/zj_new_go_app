package pkg

import (
	"fmt"
	"log"
)

// MWTNode multi way tree node.
type MWTNode struct {
	Key       string // package.func
	Value     FuncDesc
	N         int // number of leaf nodes
	LeafNodes []*MWTNode
}

// MWTree multi way tree.
type MWTree struct {
	Head *MWTNode
}

// MWTreeBuildParam params to build multi way tree.
type MWTreeBuildParam struct {
	GoFilePath string
	PkgName    string
	FnName     string
}

// BuildReverseCallTreeFromCallMap creates reverse call multi way tree from caller relationship.
func (tree *MWTree) BuildReverseCallTreeFromCallMap(param MWTreeBuildParam, callMap map[string]CallerRelation) {
	tree.Head = &MWTNode{
		Key:       fmt.Sprintf("%s.%s", param.PkgName, param.FnName),
		Value:     FuncDesc{param.GoFilePath, param.PkgName, param.FnName},
		LeafNodes: make([]*MWTNode, 0),
	}

	nodeMap := make(map[string]struct{}) // mark whether node has been handled
	nodeList := make([]*MWTNode, 1)      // a queue
	nodeList[0] = tree.Head

	for {
		if len(nodeList) == 0 {
			return
		}

		curNode := nodeList[0]
		log.Printf("current node %+v", curNode)
		for callerName, callRelation := range callMap {
			for _, callee := range callRelation.Callees {
				if curNode.Key == fmt.Sprintf("%s.%s", callee.Package, callee.Name) {
					log.Printf("found caller:%s -> callee:%s", callerName, callee)

					key := fmt.Sprintf("%s.%s", callRelation.Caller.Package, callRelation.Caller.Name)
					if _, ok := nodeMap[key]; !ok {
						newNode := &MWTNode{
							Key:       key,
							Value:     FuncDesc{callRelation.Caller.File, callRelation.Caller.Package, callRelation.Caller.Name},
							LeafNodes: make([]*MWTNode, 0),
						}
						curNode.N++
						curNode.LeafNodes = append(curNode.LeafNodes, newNode)
						nodeList = append(nodeList, newNode)
					} else {
						// node has been handled
						nodeMap[key] = struct{}{}
					}
				}
			}
		}

		nodeList = nodeList[1:]
		// log.Printf("head %+v", head)
		log.Printf("nodeList len:%d", len(nodeList))
	}
}

// GetReverseCallRelations returns reverse call relations of multi way tree.
func (tree *MWTree) GetReverseCallRelations() []ReverseCallRelation {
	// golbal vars "relation" and "relationList" for depthTraversal recursion
	relation := ReverseCallRelation{
		Callees: make([]FuncDesc, 0),
	}
	relationList := make([]ReverseCallRelation, 0)
	depthTraversal(tree.Head, relation, &relationList)
	return relationList
}

func depthTraversal(head *MWTNode, relation ReverseCallRelation, reList *[]ReverseCallRelation) {
	relation.Callees = append(relation.Callees, head.Value)
	if head.N == 0 {
		log.Printf("found reverse callees:%+v", relation.Callees)
		*reList = append(*reList, relation)
		relation.Callees = make([]FuncDesc, 0)
	} else {
		for _, node := range head.LeafNodes {
			depthTraversal(node, relation, reList)
		}
	}
}
