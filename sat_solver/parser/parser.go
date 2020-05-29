package parser

import (
	"io"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"

	"github.com/go-sat-solver/sat_solver"
)

var (
	graphQLLexer = lexer.Must(ebnf.New(`
    Comment = ("#" | "//") { "\u0000"…"\uffff"-"\n" } .
    Ident = (alpha | "_") { "_" | alpha | digit } .
	Name = "\"" (alpha | "_") { "_" | alpha | digit } "\"" .
    Number = ("." | digit) {"." | digit} .
    Whitespace = " " | "\t" | "\n" | "\r" .
    Punct = "!"…"/" | ":"…"@" | "["…`+"\"`\""+` | "{"…"~" .

    alpha = "a"…"z" | "A"…"Z" .
    digit = "0"…"9" .
`))

	parser = participle.MustBuild(&sat_solver.Entry{},
		participle.Lexer(graphQLLexer),
		participle.Elide("Comment", "Whitespace"),
	)
)

func ParseInputFormula(inputFormula io.Reader) (error, *sat_solver.Entry) {
	ast := &sat_solver.Entry{}
	err := parser.Parse(inputFormula, ast)
	if err != nil {
		return err, nil
	}

	//repr.Println(ast)
	return nil, ast
}