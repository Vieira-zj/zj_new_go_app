package structs

import (
	"fmt"
	"log"
	"strings"
)

const (
	KeyIndexByGroup = "index_by_group"
	KeyIndexByType  = "index_by_type"
)

// Resource

type ResourceMeta struct {
	Group string
	Type  string
}

type Resource struct {
	ID   int
	Name string
	Meta ResourceMeta
}

type Resources []*Resource

func (r Resource) String() string {
	return fmt.Sprintf("id=%d,name=%s,meta.group=%s,meta.type=%s", r.ID, r.Name, r.Meta.Group, r.Meta.Type)
}

// Index 倒排索引

type IndexFunc func(*Resource) string

type IndexItem map[string]Resources

type Index struct {
	fns   map[string]IndexFunc // index_key:fn, fn returns index_value
	items map[string]IndexItem // index_key:index_value:[]resource
}

func NewDefaultIndex() Index {
	return Index{
		fns: map[string]IndexFunc{
			KeyIndexByGroup: func(r *Resource) string {
				return r.Meta.Group
			},
			KeyIndexByType: func(r *Resource) string {
				return r.Meta.Type
			},
		},
		items: map[string]IndexItem{
			KeyIndexByGroup: make(IndexItem),
			KeyIndexByType:  make(IndexItem),
		},
	}
}

func (i Index) Get(idxKey, idxValue string) (resources Resources) {
	if item, ok := i.items[idxKey]; ok {
		resources = item[idxValue]
	}
	return
}

func (i Index) Add(r *Resource) {
	for idxKey, fn := range i.fns {
		idxValue := fn(r)
		if _, ok := i.items[idxKey][idxValue]; !ok {
			i.items[idxKey][idxValue] = []*Resource{r}
		} else {
			i.items[idxKey][idxValue] = append(i.items[idxKey][idxValue], r)
		}
	}
}

func (i Index) Delete(r *Resource) error {
	sb := strings.Builder{}
	for idxKey, fn := range i.fns {
		idxValue := fn(r)
		idxToDel := -1
		for i, val := range i.items[idxKey][idxValue] {
			if r.Name == val.Name {
				idxToDel = i
				break
			}
		}
		if idxToDel == -1 {
			sb.WriteString(fmt.Sprintf("fail delete resource from index, resource not found: %s:%s name=%s\n", idxKey, idxValue, r.Name))
			continue
		}

		log.Printf("delete resource: %s:%s name=%s\n", idxKey, idxValue, r.Name)
		rs := i.items[idxKey][idxValue]
		i.items[idxKey][idxValue] = append(rs[:idxToDel], rs[idxToDel+1:]...)
	}

	if errStr := sb.String(); len(errStr) > 0 {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (i Index) Update(old, new *Resource) error {
	if err := i.Delete(old); err != nil {
		return fmt.Errorf("fail update resource from index: %v", err)
	}

	i.Add(new)
	return nil
}

func (i Index) PrettyPrint() {
	sb := strings.Builder{}
	for idxKey, item := range i.items {
		for idxValue, rs := range item {
			sb.WriteString(fmt.Sprintf("%s:%s\n", idxKey, idxValue))
			for _, r := range rs {
				sb.WriteString(fmt.Sprintf("\t%s\n", r))
			}
		}
	}
	log.Println(sb.String())
}

// IndexStore

type IndexStore struct {
	data  map[string]*Resource
	index Index // 倒排索引
}

func NewIndexStore() IndexStore {
	return IndexStore{
		data:  make(map[string]*Resource),
		index: NewDefaultIndex(),
	}
}

func (s IndexStore) Add(r Resource) error {
	if old, ok := s.data[r.Name]; ok {
		if err := s.index.Update(old, &r); err != nil {
			return fmt.Errorf("fail add index store: %v", err)
		}
	} else {
		s.index.Add(&r)
	}

	s.data[r.Name] = &r
	return nil
}

func (s IndexStore) Get(name string) (*Resource, bool) {
	r, ok := s.data[name]
	return r, ok
}

func (s IndexStore) GetByIndex(idxName, idxValue string) (resources Resources, ok bool) {
	resources = s.index.Get(idxName, idxValue)
	ok = len(resources) > 0
	return
}

func (s IndexStore) Delete(r Resource) {
	if err := s.index.Delete(&r); err != nil {
		log.Println(err.Error())
	}
	delete(s.data, r.Name)
}
