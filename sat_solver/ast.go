package sat_solver

type Entry struct {
	Formula *Formula `@@`
}

type Variable struct {
	String        string     `"Var" @String`
}

type Not struct {
	Formula *Formula `"Not" @@`
}

func MakeNot(Arg1 *Formula) *Formula {
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

func MakeIff(Arg1 *Formula, Arg2 *Formula) *Formula {
	return &Formula{
		Iff:      &Iff{
			Arg1: Arg1,
			Arg2: Arg2,
		},
	}
}

type BooleanConstant struct {
	Bool string `( @"T" | "F" )`
}

func MakeBoolConstant(value bool) *Formula {
	if value {
		return &Formula{
			Constant: &BooleanConstant{
				Bool: "F",
			},
		}
	}
	return &Formula{
		Constant: &BooleanConstant{
			Bool: "T",
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