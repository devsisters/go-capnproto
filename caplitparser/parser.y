%{
package caplitparser
func init() {
	yyErrorVerbose = true
}
%}

%union {
	Node *Node
}
%token tLISTOPEN tLISTCLOSE
%token tOBJECTOPEN tOBJECTCLOSE
%token tASSIGN tCOMMA tMINUS
%token <Node> tSTRING tINT tFLOAT tIDENT tFALSE tTRUE
%type <Node> start node objectnodes listnodes minus
%%

start
	: node {
		yylex.(*lexer).result = $1
		yylex.(*lexer).lastReduce = $1
	}
	| node tCOMMA {
		yylex.(*lexer).result = $1
		yylex.(*lexer).lastReduce = $1
	}

node
	: tOBJECTOPEN objectnodes tOBJECTCLOSE {
		$$ = $2
		yylex.(*lexer).lastReduce = $$
	}
	| tOBJECTOPEN objectnodes tCOMMA tOBJECTCLOSE {
		$$ = $2
		yylex.(*lexer).lastReduce = $$
	}
	| tLISTOPEN listnodes tLISTCLOSE {
		$$ = $2
		yylex.(*lexer).lastReduce = $$
	}
	| tLISTOPEN listnodes tCOMMA tLISTCLOSE {
		$$ = $2
		yylex.(*lexer).lastReduce = $$
	}
	| minus
	| tSTRING
	| tINT
	| tFLOAT
	| tTRUE
	| tFALSE
	| tIDENT
	;

objectnodes
	: tIDENT tASSIGN node {
		$$ = &Node{Type: NOBJECT, Val: map[string]*Node { $1.Val.(string): $3 }}
		yylex.(*lexer).lastReduce = $3
	}
	| objectnodes tCOMMA tIDENT tASSIGN node {
		$$.Val.(map[string]*Node)[$3.Val.(string)] = $5
		yylex.(*lexer).lastReduce = $5
	}
	| /* empty */ {
		$$ = &Node{Type: NOBJECT, Val: make(map[string]*Node)}
	}
	;

listnodes
	: node {
		$$ = &Node{Type: NLIST, Val: []*Node{$1}}
		yylex.(*lexer).lastReduce = $1
	}
	| listnodes tCOMMA node {
		$$.Val = append($$.Val.([]*Node), $3)
		yylex.(*lexer).lastReduce = $3
	}
	| /* empty */ {
		$$ = &Node{Type: NLIST, Val: []*Node{}}
	}
	;

minus
	: tMINUS tFLOAT {
		$$ = $2
		$$.Val = -$$.Val.(float64)
	}
	| tMINUS tINT {
		$$ = $2
		$$.Val = -$$.Val.(int64)
	}
	;
%%