package caplitparser

import (
	"fmt"
	"strings"
)

func (n *Node) String() string {
	var out []string
	if n.Type == NLIST {
		out = append(out, "[")
		var inner []string
		for _, nn := range n.Val.([]*Node) {
			inner = append(inner, nn.String())
		}
		out = append(out, strings.Join(inner, ", "))
		out = append(out, "]")
	} else if n.Type == NOBJECT {
		out = append(out, "(")
		var inner []string
		for k, nn := range n.Val.(map[string]*Node) {
			inner = append(inner, fmt.Sprintf("%v=%v", k, nn.String()))
		}
		out = append(out, strings.Join(inner, ", "))
		out = append(out, ")")
	} else if n.Type == NFLOAT {
		out = append(out, fmt.Sprintf("%v", n.Val.(float64)))
	} else if n.Type == NINT {
		out = append(out, fmt.Sprintf("%v", n.Val.(int64)))
	} else if n.Type == NSTRING {
		out = append(out, fmt.Sprintf("%v", n.Val.(string)))
	} else if n.Type == NENUM {
		out = append(out, fmt.Sprintf("%v", n.Val.(string)))
	}
	return strings.Join(out, "")
}
