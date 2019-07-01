package main

// capnp literal을 capnp object로 바꿔주는 function을 generate한다.

import (
	"C"
	"fmt"
	"io"
	"strings"
)

func (root *node) defineTypeCaplitUnmarshalFuncs(w io.Writer) {
	switch root.Which() {
	case NODE_STRUCT:
		param := map[string]interface{}{
			"name": root.name,
		}
		writeTemplate(w, `
			func New{{.name}}FromCapLit(s *C.Segment, b []byte) (p {{.name}}, err error) {
				n, err := C.ParseCapLit(b)
				if err != nil { 
					return 
				}
				p = New{{.name}}(s)
				err = p.UnmarshalCapLitNode(n)
				return
			}
			func NewRoot{{.name}}FromCapLit(s *C.Segment, b []byte) (p {{.name}}, err error) {
				n, err := C.ParseCapLit(b)
				if err != nil { 
					return 
				}
				p = NewRoot{{.name}}(s)
				err = p.UnmarshalCapLitNode(n)
				return
			};
		`, param)
		param["ensureCode"] = ensureCaplitType("sv", "n.Val", "map[string]*C.CapLitNode", "")
		param["code"] = root.caplitUnmarshalStruct(root)
		writeTemplate(w, `
			func (s {{.name}}) UnmarshalCapLitNode(n *C.CapLitNode) error {
				var err error
				{{.ensureCode}}
				{{.code}}
				return err
			};
		`, param)
	case NODE_ENUM:
		param := map[string]interface{}{
			"name": root.name,
		}
		param["ensureCode"] = ensureCaplitType("ev", "n.Val", "string", fmt.Sprintf("%s(0), ", root.name))
		writeTemplate(w, `
			func {{.name}}FromCapLit(b []byte) ({{.name}}, error) {
				n, err := C.ParseCapLit(b)
				if err != nil {
					return {{.name}}(0), err
				}
				{{.ensureCode}}
				return {{.name}}FromString(ev), nil
			};
		`, param)
	default:
		panic(fmt.Sprintf("%s with type %v is not supported node type", root.name, root.Which()))
	}
}

func (n *node) caplitUnmarshalStruct(root *node) string {
	var sb strings.Builder
	for _, f := range n.codeOrderFields() {
		param := map[string]interface{}{
			"name": f.Name(),
		}
		param["code"] = f.caplitUnmarshal(root)
		writeTemplate(&sb, `
			if fn, ok := sv["{{.name}}"]; ok {
				_ = fn
				{{.code}}
			};
		`, param)
	}
	return sb.String()
}

func (f *Field) caplitUnmarshal(root *node) string {
	var sb strings.Builder
	param := map[string]interface{}{
		"name": title(f.Name()),
	}
	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		if fs.Type().Which() == TYPE_VOID {
			if f.DiscriminantValue() == 0xFFFF {
				// means it's not union. only union fields are needed to be set
				return ""
			}
			writeTemplate(&sb, `
				s.Set{{.name}}();
			`, param)
		} else {
			param["code"] = fs.Type().caplitUnmarshal(root)
			writeTemplate(&sb, `
				{{.code}}
				s.Set{{.name}}(t);
			`, param)
		}
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		param["ensureCode"] = ensureCaplitType("sv", "fn.Val", "map[string]*C.CapLitNode", "")
		param["code"] = n.caplitUnmarshalStruct(root)
		writeTemplate(&sb, `
			{{.ensureCode}}
			s := s.{{.name}}()
			_ = s
			{{.code}};
		`, param)
	default:
		panic(fmt.Sprintf("in %s, %v is not supported field type", root.name, f.Which()))
	}
	return sb.String()
}

func (t Type) caplitUnmarshal(root *node) string {
	switch t.Which() {
	case TYPE_UINT8:
		return ensureCaplitType("t", "fn.Val", "uint8", "")
	case TYPE_UINT16:
		return ensureCaplitType("t", "fn.Val", "uint16", "")
	case TYPE_UINT32:
		return ensureCaplitType("t", "fn.Val", "uint32", "")
	case TYPE_UINT64:
		return ensureCaplitType("t", "fn.Val", "uint64", "")
	case TYPE_INT8:
		return ensureCaplitType("t", "fn.Val", "int8", "")
	case TYPE_INT16:
		return ensureCaplitType("t", "fn.Val", "int16", "")
	case TYPE_INT32:
		return ensureCaplitType("t", "fn.Val", "int32", "")
	case TYPE_INT64:
		return ensureCaplitType("t", "fn.Val", "int64", "")
	case TYPE_FLOAT32:
		return ensureCaplitType("t", "fn.Val", "float32", "")
	case TYPE_FLOAT64:
		return ensureCaplitType("t", "fn.Val", "float64", "")
	case TYPE_BOOL:
		return ensureCaplitType("t", "fn.Val", "bool", "")
	case TYPE_TEXT:
		return ensureCaplitType("t", "fn.Val", "string", "")
	case TYPE_DATA:
		return ensureCaplitType("t", "fn.Val", "[]byte", "")
	case TYPE_ENUM:
		var sb strings.Builder
		param := map[string]interface{}{
			"scope":      findNode(t.Enum().TypeId()).remoteScope(root),
			"name":       findNode(t.Enum().TypeId()).name,
			"ensureCode": ensureCaplitType("fv", "fn.Val", "string", ""),
		}
		writeTemplate(&sb, `
			{{.ensureCode}}
			t := {{.scope}}{{.name}}FromString(fv);
		`, param)
		param["ensureCode"] = sb.String()
		return sb.String()
	case TYPE_STRUCT:
		var sb strings.Builder
		param := map[string]interface{}{
			"scope":      findNode(t.Struct().TypeId()).remoteScope(root),
			"name":       findNode(t.Struct().TypeId()).name,
			"ensureCode": ensureCaplitType("fv", "fn.Val", "string", ""),
		}
		writeTemplate(&sb, `
			t := {{.scope}}New{{.name}}(s.Segment)
			err := t.UnmarshalCapLitNode(fn)
			if err != nil {
				return err
			};
		`, param)
		return sb.String()
	case TYPE_LIST:
		return t.List().caplitUnmarshal(root)
	default:
		panic(fmt.Sprintf("in %s, %v is not supported type", root.name, t.Which()))
	}
}

func (t TypeList) caplitUnmarshal(root *node) string {
	param := map[string]interface{}{
		"param": "",
		"code":  "",
	}
	switch t.ElementType().Which() {
	case TYPE_UINT8:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "uint8", "")
	case TYPE_UINT16:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "uint16", "")
	case TYPE_UINT32:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "uint32", "")
	case TYPE_UINT64:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "uint64", "")
	case TYPE_INT8:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "int8", "")
	case TYPE_INT16:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "int16", "")
	case TYPE_INT32:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "int32", "")
	case TYPE_INT64:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "int64", "")
	case TYPE_FLOAT32:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "float32", "")
	case TYPE_FLOAT64:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "float64", "")
	case TYPE_BOOL:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "bool", "")
	case TYPE_TEXT:
		param["ensureCode"] = ensureCaplitType("ev", "en.Val", "string", "")
	case TYPE_ENUM:
		param["scope"] = findNode(t.ElementType().Enum().TypeId()).remoteScope(root)
		param["name"] = findNode(t.ElementType().Enum().TypeId()).name
		param["param"] = "s.Segment, "
		param["ensureCode"] = ensureCaplitType("sEv", "en.Val", "string", "")
		var sb strings.Builder
		writeTemplate(&sb, `
			var ev {{.scope}}{{.name}}
			{{.ensureCode}}
			ev = {{.scope}}{{.name}}FromString(sEv)
		`, param)
		param["ensureCode"] = sb.String()
	case TYPE_STRUCT:
		param["scope"] = findNode(t.ElementType().Struct().TypeId()).remoteScope(root)
		param["name"] = findNode(t.ElementType().Struct().TypeId()).name
		param["param"] = "s.Segment, "
		var sb strings.Builder
		writeTemplate(&sb, `
			ev := {{.scope}}New{{.name}}(s.Segment)
			err := ev.UnmarshalCapLitNode(en)
			if err != nil {
				return err
			};
		`, param)
		param["ensureCode"] = sb.String()
	case TYPE_LIST, TYPE_ANYPOINTER:
		return fmt.Sprintf(`t := C.PointerList{};panic("in %s, list of list' or 'list of pointer' is not supported")`, root.name)
	default:
		panic(fmt.Sprintf("in %s, %v is not supported list element type", root.name, t.ElementType().Which()))
	}
	param["newList"] = t.ElementType().NewListFuncString(root, "len(lv)")
	var sb strings.Builder
	writeTemplate(&sb, `
		lv := fn.Val.([]*C.CapLitNode)
		t := {{.newList}}
		for i, en := range lv {
			_ = en
			{{.ensureCode}}
			t.Set(i, ev)
		};
	`, param)
	return sb.String()
}

func ensureCaplitType(out string, target string, t string, returnWith string) string {
	g_imported["fmt"] = true
	var sb strings.Builder
	param := map[string]interface{}{
		"out":        out,
		"target":     target,
		"t":          t,
		"returnWith": returnWith,
	}
	if strings.Contains(t, "int") {
		writeTemplate(&sb, `
			var {{.out}} {{.t}}
			if temp, ok := {{.target}}.(int64); ok {
				{{.out}} = {{.t}}(temp)
			} else {
				return {{.returnWith}}fmt.Errorf("expected '{{.t}}' but didn't matched")
			}
			_ = {{.out}};
		`, param)
	} else if strings.Contains(t, "float") {
		writeTemplate(&sb, `
			var {{.out}} {{.t}}
			if temp, ok := {{.target}}.(float64); ok {
				{{.out}} = {{.t}}(temp)
			} else {
				return {{.returnWith}}fmt.Errorf("expected '{{.t}}' but didn't matched")
			}
			_ = {{.out}};
		`, param)
	} else if t == "[]byte" {
		writeTemplate(&sb, `
			var {{.out}} {{.t}}
			if temp, ok := {{.target}}.(string); ok {
				{{.out}} = {{.t}}(temp)
			} else {
				return {{.returnWith}}fmt.Errorf("expected '{{.t}}' but didn't matched")
			}
			_ = {{.out}};
		`, param)
	} else if t == "bool" || t == "map[string]*C.CapLitNode" || t == "string" {
		writeTemplate(&sb, `
			{{.out}}, ok := {{.target}}.({{.t}})
			if !ok {
				return {{.returnWith}}fmt.Errorf("expected '{{.t}}' but didn't matched")
			}
			_ = {{.out}};
		`, param)
	} else {
		panic(fmt.Sprintf("unsupported type %s", t))
	}
	return sb.String()
}
