package jsondiff

import (
	"fmt"
	"hash/maphash"
	"sort"
)

type hasher struct {
	mh maphash.Hash
}

// digest returns the object hash.
func (h *hasher) digest(val any) uint64 {
	h.mh.Reset()
	h.hash(val)
	return h.mh.Sum64()
}

func (h *hasher) hash(val any) {
	switch v := val.(type) {
	case nil:
		_ = h.mh.WriteByte('0')
	case bool:
		if v {
			_ = h.mh.WriteByte('1')
		} else {
			_ = h.mh.WriteByte('0')
		}
	case string:
		_, _ = h.mh.WriteString(v)
	case []any:
		for _, e := range v {
			h.hash(e)
		}
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			_, _ = h.mh.WriteString(k)
			h.hash(v[k])
		}
	default:
		// int numbers
		_, _ = h.mh.WriteString(fmt.Sprintf("%v", v))
	}
}
