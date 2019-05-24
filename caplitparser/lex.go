//go:generate goyacc -l -o parser.go parser.y

package caplitparser

import (
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"strconv"
)

type Node struct {
	Type NodeType
	Val  interface{}
}

type NodeType uint8

const (
	NOBJECT NodeType = iota
	NLIST
	NSTRING
	NINT
	NFLOAT
	NENUM
	NBOOL
)

type lexer struct {
	s          scanner.Scanner
	err        error
	lastReduce *Node
	lastIdent  string
	result     *Node
	int
}

func Parse(input []byte) (*Node, error) {
	p := yyNewParser()
	lexer := &lexer{}
	lexer.s = scanner.Scanner{}
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(input))
	lexer.s.Init(file, input, nil, 3)
	_ = p.Parse(lexer)
	if lexer.err != nil {
		return nil, lexer.err
	}
	return lexer.result, nil
}

func (l *lexer) Lex(lval *yySymType) int {
	_, tok, lit := l.s.Scan()
	if tok >= token.BREAK {
		lval.Node = &Node{Type: NENUM, Val: lit}
		return tIDENT
	} else {
		switch tok {
		case token.LPAREN:
			return tOBJECTOPEN
		case token.RPAREN:
			return tOBJECTCLOSE
		case token.LBRACK:
			return tLISTOPEN
		case token.RBRACK:
			return tLISTCLOSE
		case token.COMMA:
			return tCOMMA
		case token.ASSIGN:
			return tASSIGN
		case token.STRING:
			s, err := strconv.Unquote(lit)
			if err != nil {
				panic(err)
			}
			lval.Node = &Node{Type: NSTRING, Val: s}
			return tSTRING
		case token.IDENT:
			if lit == "true" {
				lval.Node = &Node{Type: NBOOL, Val: true}
				return tTRUE
			} else if lit == "false" {
				lval.Node = &Node{Type: NBOOL, Val: false}
				return tFALSE
			} else {
				lval.Node = &Node{Type: NENUM, Val: lit}
				l.lastIdent = lit
				return tIDENT
			}
		case token.INT:
			v, err := strconv.ParseInt(lit, 10, 64)
			if err != nil {
				return 0
			}
			lval.Node = &Node{Type: NINT, Val: v}
			return tINT
		case token.FLOAT:
			v, err := strconv.ParseFloat(lit, 64)
			if err != nil {
				return 0
			}
			lval.Node = &Node{Type: NFLOAT, Val: v}
			return tFLOAT
		case token.SUB:
			return tMINUS
		case token.EOF:
			return 0
		default:
			panic("Unexpected token.")
			return 0
		}
	}
}

func (l *lexer) Error(s string) {
	msg := fmt.Sprintf("%s\nlast ident: '%s'", s, l.lastIdent)
	if l.lastReduce != nil {
		msg += ", last reduced node: "
		msg += l.lastReduce.String()
	}
	msg += "\n"
	l.err = errors.New(msg)
}
