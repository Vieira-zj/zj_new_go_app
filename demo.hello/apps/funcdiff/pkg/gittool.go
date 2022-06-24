package pkg

const (
	fileDiffTypeAdd    = "add"
	fileDiffTypeDelete = "delete"
	fileDiffTypeUpdate = "update"
	fileDiffTypeRename = "rename"
)

// DiffFile .
type DiffFile struct {
	Name   string
	Rename string
	DType  string
}

func getDiffFilesByCommits(srcHash, dstHash string) []DiffFile {
	// TODO:
	return nil
}
