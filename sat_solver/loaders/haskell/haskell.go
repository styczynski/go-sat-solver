package haskell

import (
	"io"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"

	"github.com/go-sat-solver/sat_solver"
	solver "github.com/go-sat-solver/sat_solver/loaders"
)

type HaskellLoaderFactory struct {}

type HaskellLoader struct {}

func (hlf *HaskellLoaderFactory) CreateLoader(context *sat_solver.SATContext) solver.Loader {
	return HaskellLoader{}
}

func (hlf *HaskellLoaderFactory) GetName() string {
	return "haskell"
}

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

func (loader HaskellLoader) Load(inputFormula io.Reader, context *sat_solver.SATContext) (error, solver.LoadedFormula) {
	ast := &sat_solver.Entry{}
	err := parser.Parse(inputFormula, ast)
	if err != nil {
		return err, nil
	}
	return nil, ast
}

func init() {
	solver.RegisterLoaderFactory(&HaskellLoaderFactory{})
}