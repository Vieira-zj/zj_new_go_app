package pkg

import (
	"context"
	"fmt"
	"time"

	pkg "demo.hello/gocplugin/pkg"
)

const (
	fileDiffTypeAdd    = "add"
	fileDiffTypeDelete = "delete"
	fileDiffTypeUpdate = "update"
	fileDiffTypeRename = "rename"
)

var isDebug bool

// FileDiffEntry .
type FileDiffEntry struct {
	DType   string `json:"type"`
	SrcName string `json:"src_name"`
	DstName string `json:"dst_name"`
}

// getDiffFilesByCommits: 1.syncs and checkouts to branch; 2.returns diff files bewtween commits of branch.
func getDiffFilesByCommits(repoPath, branch string, srcHash, dstHash string) ([]FileDiffEntry, error) {
	repo := pkg.NewGitRepo(repoPath)
	if !isDebug {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if _, err := repo.Pull(ctx, branch); err != nil {
			return nil, err
		}
	}

	srcCommit, err := repo.GetCommit(srcHash)
	if err != nil {
		return nil, err
	}
	dstCommit, err := repo.GetCommit(dstHash)
	if err != nil {
		return nil, err
	}
	patch, err := srcCommit.Patch(dstCommit)
	if err != nil {
		return nil, err
	}

	patches := patch.FilePatches()
	diffs := make([]FileDiffEntry, 0, len(patches))
	for _, fpatch := range patches {
		from, to := fpatch.Files()
		entry := FileDiffEntry{}
		if from == nil && to != nil {
			entry.DType = fileDiffTypeAdd
			entry.DstName = to.Path()
		} else if from != nil && to == nil {
			entry.DType = fileDiffTypeDelete
			entry.SrcName = from.Path()
		} else if from != nil && to != nil {
			if from.Path() == to.Path() {
				entry.DType = fileDiffTypeUpdate
			} else {
				entry.DType = fileDiffTypeRename
			}
			entry.SrcName = from.Path()
			entry.DstName = to.Path()
		} else {
			return nil, fmt.Errorf("invalid patch: from and to files are nil")
		}
		diffs = append(diffs, entry)
	}

	return diffs, nil
}
