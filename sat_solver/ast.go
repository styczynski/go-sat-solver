package sat_solver

type Entry struct {
	Formula *Formula `@@`
}

type Variable struct {
	String        *string     `"(" "Var" @String ")"`
}

type Not struct {
	Formula *Formula `"(" "Not" @@ ")"`
}

type And struct {
	Arg1 *Formula `"(" "And" @@`
	Arg2 *Formula ` @@ ")"`
}

type Or struct {
	Arg1 *Formula `"(" "Or" @@`
	Arg2 *Formula ` @@ ")"`
}

type Implies struct {
	Arg1 *Formula `"(" "Implies" @@`
	Arg2 *Formula ` @@ ")"`
}

type Iff struct {
	Arg1 *Formula `"(" "Iff" @@`
	Arg2 *Formula ` @@ ")"`
}

type BooleanConstant struct {
	Bool *bool `( @"T" | "F" )`
}

type Formula struct {
	Constant *BooleanConstant `  @@`
	Variable *Variable        ` | @@`
	Not      *Not             ` | @@`
	And      *And             ` | @@`
	Or       *Or              ` | @@`
	Implies  *Implies         ` | @@`
	Iff      *Iff             ` | @@`
}