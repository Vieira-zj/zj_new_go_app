package structs

import (
	"fmt"
	"sort"
)

// Refer: https://github.com/qiniu/goc/blob/master/pkg/cover/internal/tool/edit.go

// An edit records a single text modification: change the bytes in [start,end) to new.
type edit struct {
	start int
	end   int
	text  string
}

// An edits is a list of edits that is sortable by start offset, breaking ties by end offset.
type edits []edit

func (x edits) Len() int { return len(x) }

func (x edits) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x edits) Less(i, j int) bool {
	if x[i].start != x[j].start {
		return x[i].start < x[j].start
	}
	return x[i].end < x[j].end
}

// An EditBuffer is a queue of edits to apply to a given byte slice.
type EditBuffer struct {
	old   []byte
	queue edits
}

func NewEditBuffer(data []byte) *EditBuffer {
	return &EditBuffer{old: data}
}

func (b *EditBuffer) Insert(pos int, new string) error {
	if pos < 0 || pos > len(b.old) {
		return fmt.Errorf("invalid edit position")
	}
	b.queue = append(b.queue, edit{pos, pos, new})
	return nil
}

func (b *EditBuffer) Delete(start, end int) error {
	if end < start || start < 0 || end > len(b.old) {
		return fmt.Errorf("invalid edit position")
	}
	b.queue = append(b.queue, edit{start, end, ""})
	return nil
}

func (b *EditBuffer) Replace(start, end int, new string) error {
	if end < start || start < 0 || end > len(b.old) {
		return fmt.Errorf("invalid edit position")
	}
	b.queue = append(b.queue, edit{start, end, new})
	return nil
}

// Bytes returns a new byte slice containing the original data with the queued edits applied.
func (b *EditBuffer) Bytes() ([]byte, error) {
	sort.Stable(b.queue)

	new := make([]byte, 0)
	offset := 0 // offset for old bytes
	for i, edt := range b.queue {
		if edt.start < offset {
			lastedt := b.queue[i-1]
			return nil, fmt.Errorf("overlapping edits: [%d,%d)->%q, [%d,%d)->%q",
				lastedt.start, lastedt.end, lastedt.text, edt.start, edt.end, edt.text)
		}

		new = append(new, b.old[offset:edt.start]...)
		new = append(new, edt.text...)
		offset = edt.end
	}
	new = append(new, b.old[offset:]...)

	return new, nil
}

func (b *EditBuffer) String() (string, error) {
	bs, err := b.Bytes()
	return string(bs), err
}
