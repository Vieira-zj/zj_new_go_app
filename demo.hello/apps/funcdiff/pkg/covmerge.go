package pkg

// BlockCovInfo .
type BlockCovInfo struct {
}

// FuncCovInfo .
type FuncCovInfo struct {
	FuncInfo
	Blocks []*BlockCovInfo
}
