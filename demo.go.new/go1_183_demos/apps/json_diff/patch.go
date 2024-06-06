package jsondiff

import (
	"encoding/json"
	"strings"
)

const (
	OpAdd     = "add"
	OpRemove  = "remove"
	OpReplace = "replace"
)

type patch struct {
	SrcValue any    `json:"src_value"`
	DstValue any    `json:"dst_value"`
	Op       string `json:"op"`
	Path     string `json:"path"`
}

type patches []patch

func (p *patches) append(src, dst any, op, path string) {
	*p = append(*p, patch{
		SrcValue: src,
		DstValue: dst,
		Op:       op,
		Path:     path,
	})
}

func (p patches) len() int {
	return len(p)
}

func (p patches) string() string {
	sb := strings.Builder{}
	for _, op := range p {
		b, err := json.Marshal(op)
		if err != nil {
			b = []byte("invalid operation")
		}
		sb.Write(b)
		sb.WriteByte('\n')

	}

	return sb.String()
}
