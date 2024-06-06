package jsondiff

import (
	"reflect"
)

type Differ struct {
	opts    options
	cursor  cursor
	patches patches
	hasher  hasher
}

func NewDiffer(optsFns ...optsFunc) Differ {
	opts := options{
		ignores:     make(map[string]struct{}),
		sliceOrders: make(map[string]struct{}),
	}
	for _, fn := range optsFns {
		fn(&opts)
	}

	return Differ{opts: opts}
}

func (d *Differ) Reset() {
	d.patches = d.patches[:0]
	d.cursor.reset()
}

func (d *Differ) Patches() patches {
	return d.patches
}

func (d *Differ) Compare(src, dst any) {
	d.diff(d.cursor, src, dst)
}

func (d *Differ) diff(cur cursor, src, dst any) {
	if src == nil && dst == nil {
		return
	}
	if d.isIgnore(cur) {
		return
	}

	if !isSameKind(src, dst) {
		d.patches.append(src, dst, OpReplace, cur.string())
		return
	}
	if d.fastEqual(src, dst) {
		return
	}

	switch src.(type) {
	case []any:
		d.compareSlice(cur, src.([]any), dst.([]any))
	case map[string]any:
		d.compareMap(cur, src.(map[string]any), dst.(map[string]any))
	default:
		if !d.fastEqual(src, dst) {
			d.patches.append(src, dst, OpReplace, cur.string())
		}
	}
}

func (d Differ) isIgnore(cur cursor) bool {
	if len(d.opts.ignores) == 0 || cur.isRoot() {
		return false
	}

	_, ok := d.opts.ignores[cur.string()]
	return ok
}

func (d Differ) fastEqual(src, dst any) bool {
	return reflect.DeepEqual(src, dst)
}

func (d *Differ) compareSlice(cur cursor, src, dst []any) {
	if _, ok := d.opts.sliceOrders[cur.string()]; ok {
		if d.orderAndCompareSlice(src, dst) {
			return
		}
	}

	minl := min(len(src), len(dst))
	if len(src) > len(dst) {
		for i := minl; i < len(src); i++ {
			cur.appendIndex(i)
			if !d.isIgnore(cur) {
				d.patches.append(src[i], nil, OpRemove, cur.string())
			}
			cur.rollback()
		}
	}

	if len(src) < len(dst) {
		for i := minl; i < len(dst); i++ {
			cur.appendIndex(i)
			if !d.isIgnore(cur) {
				d.patches.append(nil, dst[i], OpAdd, cur.string())
			}
			cur.rollback()
		}
	}

	for i := 0; i < minl; i++ {
		cur.appendIndex(i)
		d.diff(cur, src[i], dst[i])
		cur.rollback()
	}
}

func (d Differ) orderAndCompareSlice(src, dst []any) bool {
	if len(src) != len(dst) {
		return false
	}

	valSet := make(map[uint64]struct{}, len(src))
	for _, val := range src {
		key := d.hasher.digest(val)
		valSet[key] = struct{}{}
	}

	for _, val := range dst {
		key := d.hasher.digest(val)
		if _, ok := valSet[key]; !ok {
			return false
		}
	}
	return true
}

func (d *Differ) compareMap(cur cursor, src, dst map[string]any) {
	keySet := make(map[string]uint8, max(len(src), len(dst)))

	for k := range src {
		keySet[k] |= 1 << 0
	}
	for k := range dst {
		keySet[k] |= 1 << 1
	}

	for k, v := range keySet {
		cur.appendKey(k)
		inOld := v&(1<<0) != 0
		inNew := v&(1<<1) != 0

		switch {
		case inOld && inNew:
			d.diff(cur, src[k], dst[k])
		case inOld && !inNew:
			if !d.isIgnore(cur) {
				d.patches.append(src[k], dst[k], OpRemove, cur.string())
			}
		case !inOld && inNew:
			if !d.isIgnore(cur) {
				d.patches.append(src[k], dst[k], OpAdd, cur.string())
			}
		}
		cur.rollback()
	}
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
