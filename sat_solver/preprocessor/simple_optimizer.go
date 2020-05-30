package preprocessor

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-sat-solver/sat_solver"
)

func hashVarID(varID int64) int64 {
	if varID < 0 {
		return (1 >> int64(-varID) % 63)
	}
	return (1 >> int64(varID) % 63)
}

type UnsatReasonUP struct {}

func NewUnsatReasonUP() *UnsatReasonUP {
	return &UnsatReasonUP{}
}

func (reason *UnsatReasonUP) Describe() string {
	return fmt.Sprintf("Unit propagation cannot propagate variable with T or F.")
}

type UnsatReasonStrengthening struct {
	clause string
	varName string
}

func NewUnsatReasonStrengthening(clause *Clause, varID int64, opt *SimpleOptimizer) *UnsatReasonStrengthening {
	return &UnsatReasonStrengthening{
		clause: clause.String(opt),
		varName: opt.vars.Reverse(varID),
	}
}

func (reason *UnsatReasonStrengthening) Describe() string {
	return fmt.Sprintf("Strengthening clause %s by variable %s produced an empty clause.", reason.clause, reason.varName)
}

type Clause struct {
	hash int64
	vars map[int64]struct{}
	isDeleted bool
}

func (c *Clause) negateClauseVar(varIDToNegate int64) *Clause {
	newVars := map[int64]struct{}{}
	for varID := range c.vars {
		if varID == varIDToNegate || varID == -varIDToNegate {
			newVars[-varID] = struct{}{}
		} else {
			newVars[varID] = struct{}{}
		}
	}
	return &Clause{
		hash:  c.hash,
		vars:  newVars,
	}
}

func (c *Clause) String(bve *SimpleOptimizer) string {
	strs := []string{}
	for v := range c.vars {
		strs = append(strs, bve.vars.Reverse(v))
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, " v "))
}

func (c *Clause) Rehash() {
	hashVal := int64(0)
	for v := range c.vars {
		hashVal = hashVal | hashVarID(v)
	}
	c.hash = hashVal
}

type SimpleOptimizer struct {
	clauses map[*Clause]struct{}
	occur map[int64]map[*Clause]struct{}
	singular map[*Clause]struct{}
	occurBi map[int64]map[*Clause]struct{}

	added map[*Clause]struct{}
	touched map[int64]struct{}
	strenghtened map[*Clause]struct{}

	vars *sat_solver.SATVariableMapping
}

func (opt *SimpleOptimizer) Formula() *sat_solver.SATFormula {
	newFormula := sat_solver.CNFFormula{
		Variables: make([][]int64, len(opt.clauses)),
	}
	i := 0
	for c := range opt.clauses {
		newClause := make([]int64, len(c.vars))
		j := 0
		for v := range c.vars {
			newClause[j] = v
			j++
		}
		newFormula.Variables[i] = newClause
		i++
	}
	return sat_solver.NewSATFormula(&newFormula, opt.vars)
}

func (opt *SimpleOptimizer) notEqual(clause *Clause, clause2 *Clause) bool {
	if clause == clause2 {
		return false
	}
	if clause.hash != clause2.hash {
		return true
	}
	if len(clause.vars) != len(clause2.vars) {
		return true
	}
	for v := range clause.vars {
		if _, ok := clause2.vars[v]; !ok {
			return true
		}
	}
	return false
}

func (opt *SimpleOptimizer) subset(clause *Clause, clause2 *Clause) bool {
	//fmt.Printf("  Is %s subset of %s?", clause.String(opt), clause2.String(opt))
	if (clause.hash & ^clause2.hash != 0) {
		return false
	}
	found := false
	for varA := range clause.vars {
		found = false
		for varB := range clause2.vars {
			if varA == varB {
				found = true
				break
			}
		}
		if !found {
			//fmt.Printf("  NO\n")
			return false
		}
	}
	//fmt.Printf("  YES\n")
	return true
}

func (opt *SimpleOptimizer) findSubsumed(clause *Clause) []*Clause {
	res := []*Clause{}
	pLit := int64(0)
	pMin := int64(math.MaxInt64)
	for v := range clause.vars {
		occurLen := int64(len(opt.occur[v]))
		if occurLen < pMin {
			pMin = occurLen
			pLit = v
		}
	}
	if pLit == 0 {
		return res
	}
	for cPrim := range opt.occur[pLit] {
		if !cPrim.isDeleted {
			//fmt.Printf("Check %s <%p> and %s <%p>\n", clause.String(opt), clause, cPrim.String(opt), cPrim)
			if opt.notEqual(clause, cPrim) && len(clause.vars) <= len(cPrim.vars) && opt.subset(clause, cPrim) {
				res = append(res, cPrim)
			}
		}
	}
	return res
}

func (opt *SimpleOptimizer) getAddedClauseCandidates(added *map[*Clause]struct{}, positiveSearch bool) map[*Clause]struct{} {
	for clause := range *added {
		for v := range clause.vars {
			if (positiveSearch && v > 0) {
				res := map[*Clause]struct{}{}
				for c := range opt.occur[v] {
					res[c] = struct{}{}
				}
				return res
			} else if (!positiveSearch && v < 0) {
				res := map[*Clause]struct{}{}
				for c := range opt.occur[v] {
					res[c] = struct{}{}
				}
				return res
			}
		}
	}
	return map[*Clause]struct{}{}
}

// Remove varID from clause
func (opt *SimpleOptimizer) strenghten(clause *Clause, varID int64) error {
	//fmt.Printf("Strengthen clause: %s by %s\n", clause.String(opt), opt.vars.Reverse(varID))

	len1 := len(clause.vars)

	delete(clause.vars, varID)
	delete(opt.occur[varID], clause)
	clause.Rehash()

	len2 := len(clause.vars)

	if len1 == 2 && len2 == 1 {
		for v := range clause.vars {
			delete(opt.occurBi[v], clause)
		}
	} else if len1 > 2 && len2 == 2 {
		for v := range clause.vars {
			if _, ok := opt.occurBi[v]; !ok {
				opt.occurBi[v] = map[*Clause]struct{}{}
			}
			opt.occurBi[v][clause] = struct{}{}
		}
	}

	if len1 >= 1 && len2 == 1 {
		opt.singular[clause] = struct{}{}
	}

	if len1 == 1 && len2 == 0 {
		// Reverse change
		clause.vars[varID] = struct{}{}
		return sat_solver.NewUnsatError(NewUnsatReasonStrengthening(clause, varID, opt))
	}

	opt.strenghtened[clause] = struct{}{}
	for v := range clause.vars {
		opt.touched[v] = struct{}{}
	}

	// WHAT?
	// Remove empty clause
	//if len(clause.vars) == 0 {
	//	opt.removeClause(clause)
	//}

	//fmt.Printf("Strengthen result: %s\n", clause.String(opt))
	return nil
}

func (opt *SimpleOptimizer) removeClause(clause *Clause) {
	//fmt.Printf("Remove clause: %s\n", clause.String(opt))
	clause.isDeleted = true
	delete(opt.clauses, clause)
	if len(clause.vars) == 1 {
		delete(opt.singular, clause)
	}
	for v := range opt.occur {
		delete(opt.occur[v], clause)
	}
	for v := range opt.occurBi {
		delete(opt.occurBi[v], clause)
	}

	for v := range clause.vars {
		opt.touched[v] = struct{}{}
	}
}

// Remove any clause subsumed by the first argument
func (opt *SimpleOptimizer) subsume(clause *Clause) {
	clausesToRemove := opt.findSubsumed(clause)
	for _, c := range clausesToRemove {
		opt.removeClause(c)
	}
}

func (opt *SimpleOptimizer) removeHiddenTautologies() bool {
	for v, vBiClauses := range opt.occurBi {
		for c := range vBiClauses {
			if len(c.vars) == 2 {
				for v2 := range c.vars {
					if v != v2 {
						for c2 := range opt.occurBi[-v2] {
							if c != c2 && len(c2.vars) == 2 {
								for v3 := range c2.vars {
									if v3 != -v2 {
										// We remove c and c2 and replace them with [v, v3] clause
										opt.removeClause(c2)
										delete(opt.occurBi[v2], c)
										delete(opt.occur[v2], c)
										c.vars = map[int64]struct{}{
											v: {}, v3: {},
										}
										opt.occurBi[v3][c] = struct{}{}
										opt.occur[v3][c] = struct{}{}
										c.Rehash()
										return true
									}
								}
								break
							}
						}
						break
					}
				}
				break
			}
		}
	}
	return false
}

func (opt *SimpleOptimizer) optimizeTrivialTautologies() bool {
	detectedChange := false
	for c := range opt.clauses {
		for v := range c.vars {
			if _, ok := opt.occur[-v][c]; ok {
				// Clause contains both phi and -phi
				opt.removeClause(c)
				detectedChange = true
			}
		}
	}
	return detectedChange
}

func (opt *SimpleOptimizer) performUnitPropagation() (error, bool) {
	varToRemove := int64(0)
	for c := range opt.singular {
		if len(c.vars) == 1 && !c.isDeleted {
			for varID := range c.vars {
				varToRemove = varID
				break
			}
			break
		}
	}
	if varToRemove == 0 {
		return nil, false
	}

	willTriggerUnsat := false
	for c := range opt.occur[-varToRemove] {
		if len(c.vars) == 1 {
			// Removing varToRemove will result in empty clause
			// So we will try with negated variable
			willTriggerUnsat = true
			varToRemove = -varToRemove
			break
		}
	}

	if willTriggerUnsat {
		for c := range opt.occur[-varToRemove] {
			if len(c.vars) == 1 {
				// We get empty clause for both values for the variable
				// So we throw UNSAT
				return sat_solver.NewUnsatError(NewUnsatReasonUP()), false
			}
		}
	}

	for c := range opt.occur[-varToRemove] {
		err := opt.strenghten(c, -varToRemove)
		if err != nil {
			return sat_solver.WrapError(err, "When performing unit propagation for variable %s (removing negation)", opt.vars.Reverse(varToRemove)), false
		}
	}
	for c := range opt.occur[varToRemove] {
		opt.removeClause(c)
	}

	return nil, true
}

func (opt *SimpleOptimizer) maybeEliminate(varID int64) {
	if len(opt.occur[varID]) > 10 || len(opt.occur[-varID]) > 10 {
		return // Heuristic cut-off
	}
	opt.maybeClauseDistribute(varID)
}

func (opt *SimpleOptimizer) propagateToplevel() {
	// TODO: Implement
}

/*
 * Eliminates x by clause distribution if the result has fewer clauses than the original
 * (after removing trivially satisfied clauses)
 */
func(opt *SimpleOptimizer) maybeClauseDistribute(varID int64) {
	// TODO: Implement
}

func (opt *SimpleOptimizer) cleanup() error {
	for {
		err, r1 := opt.performUnitPropagation()
		if err != nil {
			return err
		}
		r2 := opt.optimizeTrivialTautologies()
		r3 := opt.blockedClauseElimination()
		if !r1 && !r2 && !r3 {
			break
		}
	}
	return nil
}

func (opt *SimpleOptimizer) simplify() error {

	opt.singular = map[*Clause]struct{}{}
	for clause := range opt.clauses {
		if len(clause.vars) == 1 {
			opt.singular[clause] = struct{}{}
		}
	}

	/*
	 * Set of variables
	 * A variable is added to this set if it occurs in a clause being added, removed, or strengthened. Initially all variables are "touched"
	 */
	opt.touched = map[int64]struct{}{}
	for clause := range opt.clauses {
		for v := range clause.vars {
			opt.touched[v] = struct{}{}
		}
	}

	/*
	 * Set of clauses
	 * When a clause is added to the SAT problem (e.g. by variable elimination), it is also added to this set.
	 * Initially all clauses are considered "added".
	 */
	opt.added = map[*Clause]struct{}{}
	for clause := range opt.clauses {
		opt.added[clause] = struct{}{}
	}

	/*
	 * Set of clauses
	 * When a clause is strengthened (one literal is removed, either by self-subsumption or toplevel propagation)
	 * it is added to this set. Initially the set is empty.
	 */
	opt.strenghtened = map[*Clause]struct{}{}

	for {
		// Subsumption

		//fmt.Printf("Iterate added %d\n", len(opt.added))

		S0 := opt.getAddedClauseCandidates(&opt.added,  true)
		for {
			//fmt.Printf("Iterate strenghtened\n")

			S1 := opt.getAddedClauseCandidates(&opt.added, false)
			for a := range opt.added {
				S1[a] = struct{}{}
			}
			for s := range opt.strenghtened {
				S1[s] = struct{}{}
			}
			// Clear Added and Strengthened
			opt.added = map[*Clause]struct{}{}
			opt.strenghtened = map[*Clause]struct{}{}

			// Loop
			for c := range S1 {
				err := opt.selfSubsume(c)
				if err != nil {
					return err
				}
			}
			// May strengthen/remove clauses
			opt.propagateToplevel()

			if len(opt.strenghtened) == 0 {
				break
			}
		}

		//fmt.Printf("Subsuming S0\n")
		for c := range S0 {
			if !c.isDeleted {
				opt.subsume(c)
			}
		}

		// Variable elimination

		//fmt.Printf("Variable elimination loop\n")
		for {
			//fmt.Printf("Eliminate variables\n")
			S := opt.touched
			opt.touched = map[int64]struct{}{}
			for x := range S {
				opt.maybeEliminate(x)
			}
			if len(opt.touched) == 0 {
				break
			}
		}
		err := opt.cleanup()
		if err != nil {
			return err
		}

		if len(opt.added) == 0 {
			break
		}
	}

	return nil
}

func (opt *SimpleOptimizer) selfSubsume(clause *Clause) error {
	for v := range clause.vars {
		subsumedBy := opt.findSubsumed(clause.negateClauseVar(v))
		for _, cPrim := range subsumedBy {
			//fmt.Printf("<%s> subsumes %s by %s\n", clause.String(opt), cPrim.String(opt), opt.vars.Reverse(v))
			err := opt.strenghten(cPrim, -v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (opt *SimpleOptimizer) blockedClauseElimination() bool {

	changeDetected := false
	for v, clausesWithV := range opt.occur {
		// Check if clauseWithV is blocked
		for clauseWithV := range clausesWithV {
			isBlocked := true

			//debugTraceClause := []*Clause{}
			//debugTraceClauseVarID := []int64{}

			if len(opt.occur[-v]) > 0 {
				for clauseWithNegV := range opt.occur[-v] {
					// Check if clauseWithV v clauseWithNegV is tautology
					isTautology := false
					for q := range clauseWithV.vars {
						if _, ok := opt.occur[-q][clauseWithNegV]; ok && q != v {

							//debugTraceClause = append(debugTraceClause, clauseWithNegV)
							//debugTraceClauseVarID = append(debugTraceClauseVarID, q)

							isTautology = true
							break
						}
					}
					if !isTautology {
						isBlocked = false
						break
					}
				}
				if isBlocked {
					// We can remove blocked clause
					/*fmt.Printf("Clause %s is blocked on %s:\n", clauseWithV.String(opt), opt.vars.Reverse(v))
					for i, c := range debugTraceClause {
						fmt.Printf("   by %s (var %s)\n", c.String(opt), opt.vars.Reverse(debugTraceClauseVarID[i]))
					}*/
					opt.removeClause(clauseWithV)
					changeDetected = true
				}
			}
		}
	}
	return changeDetected
}

func VariableElimination(formula *sat_solver.SATFormula) (error, *sat_solver.SATFormula) {
	f := formula.Formula()
	hashVal := int64(0)

	bve := SimpleOptimizer{
		clauses:   map[*Clause]struct{}{},
		occur:     map[int64]map[*Clause]struct{}{},
		occurBi:   map[int64]map[*Clause]struct{}{},
		vars:      formula.Variables(),
	}

	for _, clause := range f.Variables {
		clauseVars := map[int64]struct{}{}
		for _, v := range clause {
			clauseVars[v] = struct{}{}
		}

		c := &Clause{
			vars: clauseVars,
			isDeleted: false,
		}
		hashVal = 0
		for _, v := range clause {
			hashVal = hashVal | hashVarID(v)
			if _, ok := bve.occur[v]; ok {
				bve.occur[v][c] = struct{}{}
			} else {
				bve.occur[v] = map[*Clause]struct{}{
					c: {},
				}
			}
			if _, ok := bve.occur[-v]; !ok {
				bve.occur[-v] = map[*Clause]struct{}{}
			}
		}
		if len(c.vars) == 2 {
			for _, v := range clause {
				if _, ok := bve.occurBi[v]; ok {
					bve.occurBi[v][c] = struct{}{}
				} else {
					bve.occurBi[v] = map[*Clause]struct{}{
						c: {},
					}
				}
				if _, ok := bve.occurBi[-v]; !ok {
					bve.occurBi[-v] = map[*Clause]struct{}{}
				}
			}
		}

		c.hash = hashVal
		bve.clauses[c] = struct{}{}
	}

	err := bve.simplify()
	if err != nil {
		if v, ok := err.(*sat_solver.UnsatError); ok {
			f := bve.Formula()
			return nil, sat_solver.NewSATFormulaShortcut(f.Formula(), f.Variables(), v)
		}
	}

	for bve.removeHiddenTautologies() {}

	return nil, bve.Formula()
}