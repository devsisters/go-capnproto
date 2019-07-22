package main

import "fmt"

func (t Type) TypeString(root *node) string {
	tName := ""
	switch t.Which() {
	case TYPE_BOOL:
		tName = "bool"
	case TYPE_FLOAT32:
		tName = "float32"
	case TYPE_FLOAT64:
		tName = "float64"
	case TYPE_INT8:
		tName = "int8"
	case TYPE_INT16:
		tName = "int16"
	case TYPE_INT32:
		tName = "int32"
	case TYPE_INT64:
		tName = "int64"
	case TYPE_UINT8:
		tName = "uint8"
	case TYPE_UINT16:
		tName = "uint16"
	case TYPE_UINT32:
		tName = "uint32"
	case TYPE_UINT64:
		tName = "uint64"
	case TYPE_TEXT:
		tName = "string"
	case TYPE_DATA:
		tName = "[]byte"
	case TYPE_STRUCT:
		tName = findNode(t.Struct().TypeId()).remoteName(root)
	case TYPE_ENUM:
		tName = findNode(t.Enum().TypeId()).remoteName(root)
	default:
		panic(fmt.Sprintf("unexpected type id %v (in %s)", t.Which(), root.name))
	}
	return tName
}

func (t Type) NewListFuncString(root *node, length string) string {
	scope := "s.Segment."
	tName := ""
	param := ""
	switch t.Which() {
	case TYPE_BOOL:
		tName = "Bit"
	case TYPE_FLOAT32:
		tName = "Float32"
	case TYPE_FLOAT64:
		tName = "Float64"
	case TYPE_INT8:
		tName = "Int8"
	case TYPE_INT16:
		tName = "Int16"
	case TYPE_INT32:
		tName = "Int32"
	case TYPE_INT64:
		tName = "Int64"
	case TYPE_UINT8:
		tName = "UInt8"
	case TYPE_UINT16:
		tName = "UInt16"
	case TYPE_UINT32:
		tName = "UInt32"
	case TYPE_UINT64:
		tName = "UInt64"
	case TYPE_TEXT:
		tName = "Text"
	case TYPE_DATA:
		tName = "Text"
	case TYPE_LIST:
		tName = "Pointer"
	case TYPE_STRUCT:
		n := findNode(t.Struct().TypeId())
		scope = n.remoteScope(root)
		tName = n.name
		param = "s.Segment, "
	case TYPE_ENUM:
		n := findNode(t.Enum().TypeId())
		scope = n.remoteScope(root)
		tName = n.name
		param = "s.Segment, "
	default:
		panic(fmt.Sprintf("unexpected type id %v (in %s)", t.Which(), root.name))
	}
	return fmt.Sprintf("%vNew%vList(%v%v)", scope, tName, param, length)
}
