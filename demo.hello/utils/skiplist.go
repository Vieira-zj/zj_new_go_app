package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

/*
Skiplist node
*/

type sNode struct {
	key   int
	value string
	depth int
	next  []*sNode
}

func newSkiplistNode(key int, value string, depth int) *sNode {
	nodes := make([]*sNode, depth)
	return &sNode{
		key:   key,
		value: value,
		depth: depth,
		next:  nodes,
	}
}

func (node *sNode) getNextNode(depth int) (*sNode, error) {
	if err := node.verifyDepth(depth); err != nil {
		return nil, err
	}
	return node.next[depth], nil
}

func (node *sNode) getNextNodeKey(depth int) (int, error) {
	if err := node.verifyDepth(depth); err != nil {
		return -1, err
	}
	return node.next[depth].key, nil
}

func (node *sNode) setNextNode(n *sNode, depth int) error {
	if err := node.verifyDepth(depth); err != nil {
		return err
	}
	node.next[depth] = n
	return nil
}

func (node *sNode) string() string {
	return fmt.Sprintf("%d:%s:%d", node.key, node.value, node.depth)
}

func (node *sNode) verifyDepth(depth int) error {
	if depth < 0 || depth > (node.depth-1) {
		return fmt.Errorf("invalid depth, available value: [0-%d]", (node.depth - 1))
	}
	return nil
}

/*
Skiplist

1. SkipList是一个实现快速查找、增删数据的数据结构，可以做到 O(logN) 复杂度的增删查。
2. SkipList引入了随机深度的机制，也就是一个节点能够拥有的指针数量是随机的。同样这种策略来保证元素尽可能分散均匀，使得不会发生数据倾斜的情况。
*/

// Skiplist .
type Skiplist struct {
	root     *sNode
	tail     *sNode
	maxDepth int
	rate     float32
	depth    int
	inf      int
}

// NewSkiplist .
func NewSkiplist(maxDepth int, rate float32) *Skiplist {
	inf := math.MaxInt32
	root := newSkiplistNode(-inf, "", maxDepth)
	tail := newSkiplistNode(inf, "", maxDepth)
	// root节点全部指向tail
	for i := 0; i < maxDepth; i++ {
		root.setNextNode(tail, i)
	}

	return &Skiplist{
		root:     root,
		tail:     tail,
		maxDepth: maxDepth,
		rate:     rate,
		depth:    1,
		inf:      inf,
	}
}

// Query .
func (s *Skiplist) Query(key int) (string, error) {
	node := s.root
	for depth := s.depth - 1; depth >= 0; depth-- {
		for key > s.getNextNodeKey(node, depth) {
			node = s.getNextNode(node, depth)
		}
		if s.getNextNodeKey(node, depth) == key {
			fmt.Println("found at depth:", depth)
			return s.getNextNode(node, depth).value, nil
		}
	}
	return "", fmt.Errorf("[%d] not found", key)
}

// Delete .
func (s *Skiplist) Delete(key int) bool {
	ret := false
	node := s.root
	for depth := s.depth - 1; depth >= 0; depth-- {
		for key > s.getNextNodeKey(node, depth) {
			node = s.getNextNode(node, depth)
		}
		if s.getNextNodeKey(node, depth) == key {
			deleteNode := s.getNextNode(node, depth)
			nextNextNode := s.getNextNode(deleteNode, depth)
			node.setNextNode(nextNextNode, depth)
			if s.getNextNodeKey(s.root, depth) == s.inf && s.depth > 1 {
				s.depth--
			}
			ret = true
		}
	}
	return ret
}

// Insert .
func (s *Skiplist) Insert(key int, value string) {
	node := s.root
	randDepth := s.getRandomDepth()
	insertNodes := make([]*sNode, randDepth)
	for depth := randDepth - 1; depth >= 0; depth-- {
		for key > s.getNextNodeKey(node, depth) {
			node = s.getNextNode(node, depth)
		}
		if s.getNextNodeKey(node, depth) == key {
			fmt.Println("update exsiting node:", s.getNextNode(node, depth).string())
			s.getNextNode(node, depth).value = value
			return
		}
		insertNodes[depth] = node
	}

	newNode := newSkiplistNode(key, value, randDepth)
	s.setDepth(randDepth)
	fmt.Println("insert new node:", newNode.string())
	for depth, node := range insertNodes {
		nextNode := s.getNextNode(node, depth)
		node.setNextNode(newNode, depth)
		newNode.setNextNode(nextNode, depth)
	}
}

// Print .
func (s *Skiplist) Print() {
	for depth := s.depth - 1; depth >= 0; depth-- {
		node := s.root
		values := make([]string, 0, 16)
		for s.getNextNodeKey(node, depth) != s.inf {
			node = s.getNextNode(node, depth)
			values = append(values, node.string())
		}
		fmt.Printf("depth->%d: %s\n", depth, strings.Join(values, ","))
	}
}

func (s *Skiplist) getRandomDepth() int {
	rand.Seed(time.Now().UnixNano())
	depth := 1
	for {
		r := rand.Float32()
		if r < s.rate || depth == s.maxDepth {
			return depth
		}
		depth++
	}
}

func (s *Skiplist) setDepth(depth int) {
	if depth > s.depth {
		s.depth = depth
	}
}

func (s *Skiplist) getNextNode(node *sNode, depth int) *sNode {
	ret, _ := node.getNextNode(depth)
	return ret
}

func (s *Skiplist) getNextNodeKey(node *sNode, depth int) int {
	key, _ := node.getNextNodeKey(depth)
	return key
}
