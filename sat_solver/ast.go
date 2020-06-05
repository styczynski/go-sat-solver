package sat_solver

import (
	"fmt"
	"strings"
)

type Entry struct {
	Formula *Formula `@@`
}

func (f *Entry) AST() *Formula {
	return f.Formula
}

func (f *Entry) CanBeConvertedToFormula() bool {
	return false
}

func (f *Entry) CanBeConvertedToAST() bool {
	return true
}

func (f *Entry) ConvertToFormula() *SATFormula {
	return nil
}

func (f *Entry) ConvertToAST() *Entry {
	return f
}

func (f *Entry) IsCNF() bool {
	return false
}

type Variable struct {
	Name string `"Var" @Name`
}

func (astNode *Variable) String() string {
	return fmt.Sprintf("Var \"%s\"", trimVarQuotes(astNode.Name))
}

func MakeVar(name string) *Formula {
	return &Formula{
		Variable: &Variable{
			Name: name,
		},
	}
}

type Not struct {
	Formula *Formula `"Not" @@`
}

func (astNode *Not) String() string {
	return fmt.Sprintf("Not (%s)", astNode.Formula.String())
}

func MakeNot(Arg1 *Formula) *Formula {
	if Arg1.Not != nil {
		return Arg1.Not.Formula
	}
	return &Formula{
		Not:      &Not{
			Formula: Arg1,
		},
	}
}

type And struct {
	Arg1 *Formula `"And" @@`
	Arg2 *Formula ` @@`
}

func (astNode *And) String() string {
	return fmt.Sprintf("And (%s) (%s)", astNode.Arg1.String(), astNode.Arg2.String())
}

func MakeAnd(Arg1 *Formula, Arg2 *Formula) *Formula {
	return &Formula{
		And:      &And{
			Arg1: Arg1,
			Arg2: Arg2,
		},
	}
}

type Or struct {
	Arg1 *Formula `"Or" @@`
	Arg2 *Formula ` @@`
}

func (astNode *Or) String() string {
	return fmt.Sprintf("Or (%s) (%s)", astNode.Arg1.String(), astNode.Arg2.String())
}

func MakeOr(Arg1 *Formula, Arg2 *Formula) *Formula {
	return &Formula{
		Or:      &Or{
			Arg1: Arg1,
			Arg2: Arg2,
		},
	}
}

type Implies struct {
	Arg1 *Formula `"Implies" @@`
	Arg2 *Formula ` @@`
}

func (astNode *Implies) String() string {
	return fmt.Sprintf("Implies (%s) (%s)", astNode.Arg1.String(), astNode.Arg2.String())
}

func MakeImplies(Arg1 *Formula, Arg2 *Formula) *Formula {
	return &Formula{
		Implies:      &Implies{
			Arg1: Arg1,
			Arg2: Arg2,
		},
	}
}

type Iff struct {
	Arg1 *Formula `"Iff" @@`
	Arg2 *Formula ` @@`
}

func (astNode *Iff) String() string {
	return fmt.Sprintf("Iff (%s) (%s)", astNode.Arg1.String(), astNode.Arg2.String())
}

func MakeIff(Arg1 *Formula, Arg2 *Formula) *Formula {
	return &Formula{
		Iff:      &Iff{
			Arg1: Arg1,
			Arg2: Arg2,
		},
	}
}

type BooleanConstant struct {
	Bool string `( @"T" | @"F" )`
}

func (astNode *BooleanConstant) String() string {
	return astNode.Bool
}

func MakeBoolConstant(value bool) *Formula {
	if value {
		return &Formula{
			Constant: &BooleanConstant{
				Bool: "T",
			},
		}
	}
	return &Formula{
		Constant: &BooleanConstant{
			Bool: "F",
		},
	}
}

type Formula struct {
	Constant *BooleanConstant ` ( @@ | "(" @@ ")" )`
	Variable *Variable        ` | ( @@ | "(" @@ ")" )`
	Not      *Not             ` | ( @@ | "(" @@ ")" )`
	And      *And             ` | ( @@ | "(" @@ ")" )`
	Or       *Or              ` | ( @@ | "(" @@ ")" )`
	Implies  *Implies         ` | ( @@ | "(" @@ ")" )`
	Iff      *Iff             ` | ( @@ | "(" @@ ")" )`
}

func (f *Formula) AST() *Formula {
	return f
}

func (astNode *Formula) String() string {
	if astNode.Constant != nil {
		return astNode.Constant.String()
	} else if astNode.Variable != nil {
		return astNode.Variable.String()
	} else if astNode.Not != nil {
		return astNode.Not.String()
	} else if astNode.And != nil {
		return astNode.And.String()
	} else if astNode.Or != nil {
		return astNode.Or.String()
	} else if astNode.Implies != nil {
		return astNode.Implies.String()
	} else if astNode.Iff != nil {
		return astNode.Iff.String()
	}

	panic(fmt.Errorf("Unknown AST node given to Formula.Name() method."))
}

func AndChainToString(clauses []*Formula) string {
	results := []string{}
	for _, clause := range clauses {
		results = append(results, fmt.Sprintf("(%s)", clause.String()))
	}
	return strings.Join(results, " ^ ")
}

func trimVarQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '[' && s[len(s)-1] == ']' {
			return "var" + s[1: len(s)-1]
		}
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
