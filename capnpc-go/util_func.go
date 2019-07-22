package main

// 자주 쓰이는 몇 가지 util function을 generate한다.

import (
	"io"
)

func (root *node) defineTypeUtilFunc(w io.Writer) {
	switch root.Which() {
	case NODE_STRUCT:
		root.utilFunc(w)
	}
}

func (n *node) utilFunc(w io.Writer) {
	writeTemplate(w, `
		func (s {{.name}}_List) FilterIndex(f func(i int, x {{.name}}) bool) []int { filtered := make([]int, 0)
			for i := 0; i < s.Len(); i++ {
				if f(i, s.At(i)) {
					filtered = append(filtered, i)
				}
			}
			return filtered
		}
		func (s {{.name}}_List) Each(f func(i int, x {{.name}}) error) error {
			for i := 0; i < s.Len(); i++ {
				err := f(i, s.At(i))
				if err!=nil {
					return err
				}
			}
			return nil
		}
		func (s {{.name}}_List) Find(f func(i int, x {{.name}}) bool) ({{.name}}, bool) {
			for i := 0; i < s.Len(); i++ {
				if f(i, s.At(i)) {
					return s.At(i), true
				}
			}
			return {{.name}}{}, false
		}
		func (s {{.name}}) Seg() *C.Segment {
			return s.Segment
		};
		func (s {{.name}}) LitName() string {
			return "{{.name}}"
		};
	`, map[string]interface{}{"name": n.name})
	for _, f := range n.codeOrderFields() {
		if f.Name() != "id" {
			continue
		}
		writeTemplate(w, `
			func (s {{.name}}_List) FindById(id {{.tName}}) ({{.name}}, bool) {
				for i := 0; i < s.Len(); i++ {
					if s.At(i).Id() == id {
						return s.At(i), true
					}
				}
				return {{.name}}{}, false
			};
		`, map[string]interface{}{"name": n.name, "tName": f.Slot().Type().TypeString(n)})
		break
	}
}
