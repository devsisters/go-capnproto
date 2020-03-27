package main

import (
	"fmt"
	"io"
	"strings"
)

// convert capnp to go native struct
func (root *node) defineGoStructFuncs(w io.Writer) {
	switch root.Which() {
	case NODE_STRUCT:
		writeTemplate(w, `
			type {{.name}}_Struct struct {
				{{.fieldCode}}
			}
			func (s {{.name}}) Struct() *{{.name}}_Struct {
				t := &{{.name}}_Struct{}
				{{.makeCode}}
				return t
			}
			func (s {{.name}}) LoadStruct(t *{{.name}}_Struct) {
				if t == nil {
					return
				}
				{{.loadCode}}
			};
			func (s *{{.name}}_Struct) Copy() *{{.name}}_Struct {
				t := &{{.name}}_Struct{}
				{{.copyCode}}
				return t
			};
			func (s *{{.name}}_Struct) Capnp(seg *C.Segment) {{.name}} {
				if seg == nil {
					seg = C.NewBuffer(nil)
				}
				t := New{{.name}}(seg)
				t.LoadStruct(s)
				return t
			};
			func (s *{{.name}}_Struct) RootCapnp(seg *C.Segment) {{.name}} {
				if seg == nil {
					seg = C.NewBuffer(nil)
				}
				t := NewRoot{{.name}}(seg)
				t.LoadStruct(s)
				return t
			};
		`, map[string]interface{}{
			"name": root.name,
			"fieldCode": root.goStructFuncField(root),
			"makeCode": root.goStructFuncMake(root),
			"loadCode": root.goStructFuncLoad(root),
			"copyCode": root.goStructFuncCopy(root),
		})
	default:
		panic(fmt.Sprintf("%s with type %v is not supported node type", root.name, root.Which()))
	}
}

func (n *node) goStructFuncField(root *node) string {
	var sb strings.Builder
	if n.Struct().DiscriminantCount() > 0 {
		writeTemplate(&sb, `
			Which {{.name}}_Which;
		`, map[string]interface{}{
			"name": n.name,
		})
	}
	for _, f := range n.codeOrderFields() {
		sb.WriteString(f.goStructFuncField(root))
	}
	return sb.String()
}

func (f *Field) goStructFuncField(root *node) string {
	var sb strings.Builder
	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		if fs.Type().Which() != TYPE_VOID {
			writeTemplate(&sb, `
				{{.name}} {{.tName}};
			`, map[string]interface{}{
				"name": title(f.Name()),
				"tName": f.Slot().Type().goStructFuncField(root),
			})
		}
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		writeTemplate(&sb, `
				{{.name}} struct {
					{{.fieldCode}}
				};
			`, map[string]interface{}{
			"name": title(f.Name()),
			"fieldCode": n.goStructFuncField(root),
		})
	default:
		panic(fmt.Sprintf("in %s, %v is not supported field type", root.name, f.Which()))
	}
	return sb.String()
}

func (t Type) goStructFuncField(root *node) string {
	var sb strings.Builder
	switch t.Which() {
	case TYPE_BOOL,
		TYPE_UINT8,
		TYPE_UINT16,
		TYPE_UINT32,
		TYPE_UINT64,
		TYPE_INT8,
		TYPE_INT16,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_FLOAT32,
		TYPE_FLOAT64,
		TYPE_TEXT,
		TYPE_DATA,
		TYPE_ENUM:
		return t.TypeString(root)
	case TYPE_STRUCT:
		var sb strings.Builder
		writeTemplate(&sb, `
			*{{.name}}_Struct
		`, map[string]interface{}{
			"name": t.TypeString(root),
		})
		return sb.String()
	case TYPE_LIST:
		var sb strings.Builder
		writeTemplate(&sb, `
			[]{{.elemType}}
		`, map[string]interface{}{
			"elemType": t.List().ElementType().goStructFuncField(root),
		})
		return sb.String()
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.Which()))
	}
	return sb.String()
}

func (n *node) goStructFuncMake(root *node) string {
	var sb strings.Builder
	writeTemplate(&sb, `
		{{if .hasUnion}}t.Which = s.Which(){{end}};
	`, map[string]interface{}{
		"name": n.name,
		"hasUnion": n.Struct().DiscriminantCount() > 0,
	})
	for _, f := range n.codeOrderFields() {
		sb.WriteString(f.goStructFuncMake(root, n.name))
	}
	return sb.String()
}

func (f *Field) goStructFuncMake(root *node, structName string) string {
	var sb strings.Builder
	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		if fs.Type().Which() == TYPE_VOID {
			return ""
		}
		if f.DiscriminantValue() == 0xFFFF {
			writeTemplate(&sb, `
				{{.code}};
			`, map[string]interface{}{
				"code": fs.Type().goStructFuncMake(root, f.Name()),
			})
		} else {
			writeTemplate(&sb, `
				if t.Which == {{.structName}}_{{.name}} {
					{{.code}}
				};
			`, map[string]interface{}{
				"structName": strings.ToUpper(structName),
				"name": strings.ToUpper(f.Name()),
				"code": f.Slot().Type().goStructFuncMake(root, f.Name()),
			})
		}
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		writeTemplate(&sb, `
			{
				t := &t.{{.name}}
				_ = t
				s := s.{{.name}}()
				_ = s
				{{.code}}
			};
		`, map[string]interface{}{
			"name": title(f.Name()),
			"code": n.goStructFuncMake(root),
		})
	default:
		panic(fmt.Sprintf("in %s, %v is not supported field type", root.name, f.Which()))
	}
	return sb.String()
}

func (t Type) goStructFuncMake(root *node, fieldName string) string {
	var sb strings.Builder
	switch t.Which() {
	case TYPE_BOOL,
			TYPE_UINT8,
			TYPE_UINT16,
			TYPE_UINT32,
			TYPE_UINT64,
			TYPE_INT8,
			TYPE_INT16,
			TYPE_INT32,
			TYPE_INT64,
			TYPE_FLOAT32,
			TYPE_FLOAT64,
			TYPE_TEXT,
			TYPE_DATA,
			TYPE_ENUM:
		writeTemplate(&sb, `
			t.{{.name}} = s.{{.name}}();
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_STRUCT:
		writeTemplate(&sb, `
			{
				t.{{.name}} = s.{{.name}}().Struct();
			};
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_LIST:
		writeTemplate(&sb, `
			for i := 0; i < s.{{.name}}().Len(); i++ {
				{{.code}}
			};
		`, map[string]interface{}{
			"name": title(fieldName),
			"code": t.List().goStructFuncMake(root, fieldName),
		})
		return sb.String()
	case TYPE_VOID:
		return ""
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.Which()))
	}
	return sb.String()
}

func (t TypeList) goStructFuncMake(root *node, fieldName string) string {
	var sb strings.Builder
	switch t.ElementType().Which() {
	case TYPE_BOOL,
		TYPE_UINT8,
		TYPE_UINT16,
		TYPE_UINT32,
		TYPE_UINT64,
		TYPE_INT8,
		TYPE_INT16,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_FLOAT32,
		TYPE_FLOAT64,
		TYPE_TEXT,
		TYPE_DATA,
		TYPE_ENUM:
		writeTemplate(&sb, `
			t.{{.name}} = append(t.{{.name}}, s.{{.name}}().At(i));
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_STRUCT:
		writeTemplate(&sb, `
			elem := s.{{.name}}().At(i).Struct()
			t.{{.name}} = append(t.{{.name}}, elem);
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_LIST:
		writeTemplate(&sb, `
			t.{{.name}} = nil
			panic("List of List not supported now")
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.ElementType().Which()))
	}
	return sb.String()
}

func (n *node) goStructFuncLoad(root *node) string {
	var sb strings.Builder
	for _, f := range n.codeOrderFields() {
		sb.WriteString(f.goStructFuncLoad(root, n.name))
	}
	return sb.String()
}

func (f *Field) goStructFuncLoad(root *node, structName string) string {
	var sb strings.Builder
	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		if f.DiscriminantValue() == 0xFFFF {
			if fs.Type().Which() != TYPE_VOID {
				writeTemplate(&sb, `
					{{.code}};
				`, map[string]interface{}{
					"code": fs.Type().goStructFuncLoad(root, f.Name()),
				})
			}
		} else {
			writeTemplate(&sb, `
				if t.Which == {{.structName}}_{{.name}} {
					{{.code}}
				};
			`, map[string]interface{}{
				"structName": strings.ToUpper(structName),
				"name": strings.ToUpper(f.Name()),
				"code": f.Slot().Type().goStructFuncLoad(root, f.Name()),
			})
		}
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		writeTemplate(&sb, `
			{
				t := t.{{.name}}
				_ = t
				s := s.{{.name}}()
				_ = s
				{{.code}}
			};
		`, map[string]interface{}{
			"name": title(f.Name()),
			"code": n.goStructFuncLoad(root),
		})
	default:
		panic(fmt.Sprintf("in %s, %v is not supported field type", root.name, f.Which()))
	}
	return sb.String()
}

func (t Type) goStructFuncLoad(root *node, fieldName string) string {
	var sb strings.Builder
	switch t.Which() {
	case TYPE_BOOL,
		TYPE_UINT8,
		TYPE_UINT16,
		TYPE_UINT32,
		TYPE_UINT64,
		TYPE_INT8,
		TYPE_INT16,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_FLOAT32,
		TYPE_FLOAT64,
		TYPE_TEXT,
		TYPE_DATA,
		TYPE_ENUM:
		writeTemplate(&sb, `
			s.Set{{.name}}(t.{{.name}});
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_VOID:
		writeTemplate(&sb, `
			s.Set{{.name}}();
		`, map[string]interface{}{
			"name": title(fieldName),
		})
	case TYPE_STRUCT:
		tid := t.Struct().TypeId()
		n := findNode(tid)
		writeTemplate(&sb, `
			{
				p := {{.elemScope}}New{{.elemName}}(s.Segment)
				p.LoadStruct(t.{{.name}})
				s.Set{{.name}}(p)
			};
		`, map[string]interface{}{
			"name": title(fieldName),
			"elemScope": n.remoteScope(root),
			"elemName": n.name,
		})
		return sb.String()
	case TYPE_LIST:
		writeTemplate(&sb, `
			s.Set{{.name}}({{.newList}})
			for i := 0; i < len(t.{{.name}}); i++ {
				{{.code}}
			};
		`, map[string]interface{}{
			"newList": t.List().ElementType().NewListFuncString(root, fmt.Sprintf("len(t.%v)", title(fieldName))),
			"name": title(fieldName),
			"code": t.List().goStructFuncLoad(root, fieldName),
		})
		return sb.String()
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.Which()))
	}
	return sb.String()
}

func (t TypeList) goStructFuncLoad(root *node, fieldName string) string {
	var sb strings.Builder
	switch t.ElementType().Which() {
	case TYPE_BOOL,
		TYPE_UINT8,
		TYPE_UINT16,
		TYPE_UINT32,
		TYPE_UINT64,
		TYPE_INT8,
		TYPE_INT16,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_FLOAT32,
		TYPE_FLOAT64,
		TYPE_TEXT,
		TYPE_DATA,
		TYPE_ENUM:
			t.ElementType().TypeString(root)
		writeTemplate(&sb, `
			s.{{.name}}().Set(i, t.{{.name}}[i]);
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_STRUCT:
		writeTemplate(&sb, `
			s.{{.name}}().Set(i, t.{{.name}}[i].Capnp(s.Segment));
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_LIST:
		writeTemplate(&sb, `
			panic("List of List not supported now")
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.ElementType().Which()))
	}
	return sb.String()
}

func (n *node) goStructFuncCopy(root *node) string {
	var sb strings.Builder
	writeTemplate(&sb, `
		{{if .hasUnion}}t.Which = s.Which{{end}};
	`, map[string]interface{}{
		"name": n.name,
		"hasUnion": n.Struct().DiscriminantCount() > 0,
	})
	for _, f := range n.codeOrderFields() {
		sb.WriteString(f.goStructFuncCopy(root, n.name))
	}
	return sb.String()
}

func (f *Field) goStructFuncCopy(root *node, structName string) string {
	var sb strings.Builder
	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		if fs.Type().Which() == TYPE_VOID {
			return ""
		}
		writeTemplate(&sb, `
			{{.code}};
		`, map[string]interface{}{
			"code": fs.Type().goStructFuncCopy(root, f.Name()),
		})
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		writeTemplate(&sb, `
			{
				t := &t.{{.name}}
				_ = t
				s := &s.{{.name}}
				_ = s
				{{.code}}
			};
		`, map[string]interface{}{
			"name": title(f.Name()),
			"code": n.goStructFuncCopy(root),
		})
	default:
		panic(fmt.Sprintf("in %s, %v is not supported field type", root.name, f.Which()))
	}
	return sb.String()
}

func (t Type) goStructFuncCopy(root *node, fieldName string) string {
	var sb strings.Builder
	switch t.Which() {
	case TYPE_BOOL,
		TYPE_UINT8,
		TYPE_UINT16,
		TYPE_UINT32,
		TYPE_UINT64,
		TYPE_INT8,
		TYPE_INT16,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_FLOAT32,
		TYPE_FLOAT64,
		TYPE_TEXT,
		TYPE_DATA,
		TYPE_ENUM:
		writeTemplate(&sb, `
			t.{{.name}} = s.{{.name}}
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_VOID:
	case TYPE_STRUCT:
		writeTemplate(&sb, `
			if s.{{.name}} != nil {
				t.{{.name}} = s.{{.name}}.Copy()
			};
		`, map[string]interface{}{
			"name": title(fieldName),
		})
		return sb.String()
	case TYPE_LIST:
		if t.List().ElementType().Which() == TYPE_STRUCT {
			writeTemplate(&sb, `
				for _, e := range s.{{.name}} {
                    if e != nil {
						t.{{.name}} = append(t.{{.name}}, e.Copy())
                    } else {
						t.{{.name}} = append(t.{{.name}}, nil)
					}
				};
			`, map[string]interface{}{
				"name": title(fieldName),
			})
		} else {
			writeTemplate(&sb, `
				for _, e := range s.{{.name}} {
					t.{{.name}} = append(t.{{.name}}, e)
				};
			`, map[string]interface{}{
				"name": title(fieldName),
			})
		}
		return sb.String()
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.Which()))
	}
	return sb.String()
}
