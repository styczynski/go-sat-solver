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

type Clause struct {
	hash int64
	vars map[int64]struct{}
	isDeleted bool
}

func (c *Clause) String(bve *BVEContext) string {
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

type BVEContext struct {
	clauses map[*Clause]struct{}
	occur map[int64]map[*Clause]struct{}

	added map[*Clause]struct{}
	touched map[int64]struct{}
	strenghtened map[*Clause]struct{}

	vars *sat_solver.SATVariableMapping
}

func notEqual(clause *Clause, clause2 *Clause, bve *BVEContext) bool {
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

func subset(clause *Clause, clause2 *Clause, bve *BVEContext) bool {
	//fmt.Printf("  Is %s subset of %s?", clause.String(bve), clause2.String(bve))
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

func findSubsumed(clause *Clause, bve *BVEContext) []*Clause {
	res := []*Clause{}
	pLit := int64(0)
	pMin := int64(math.MaxInt64)
	for v := range clause.vars {
		occurLen := int64(len(bve.occur[v]))
		if occurLen < pMin {
			pMin = occurLen
			pLit = v
		}
	}
	if pLit == 0 {
		return res
	}
	for cPrim := range bve.occur[pLit] {
		if !cPrim.isDeleted {
			//fmt.Printf("Check %s <%p> and %s <%p>\n", clause.String(bve), clause, cPrim.String(bve), cPrim)
			if notEqual(clause, cPrim, bve) && len(clause.vars) <= len(cPrim.vars) && subset(clause, cPrim, bve) {
				res = append(res, cPrim)
			}
		}
	}
	return res
}

func getAddedClauseCandidates(added *map[*Clause]struct{}, positiveSearch bool, bve *BVEContext) map[*Clause]struct{} {
	for clause := range *added {
		for v := range clause.vars {
			if (positiveSearch && v > 0) {
				res := map[*Clause]struct{}{}
				for c := range bve.occur[v] {
					res[c] = struct{}{}
				}
				return res
			} else if (!positiveSearch && v < 0) {
				res := map[*Clause]struct{}{}
				for c := range bve.occur[v] {
					res[c] = struct{}{}
				}
				return res
			}
		}
	}
	return map[*Clause]struct{}{}
}

// Remove varID from clause
func strenghten(clause *Clause, varID int64, bve *BVEContext) {
	//fmt.Printf("Strengthen clause: %s by %s\n", clause.String(bve), bve.vars.Reverse(varID))

	delete(clause.vars, varID)
	delete(bve.occur[varID], clause)
	clause.Rehash()

	bve.strenghtened[clause] = struct{}{}
	for v := range clause.vars {
		bve.touched[v] = struct{}{}
	}

	// Remove empty clause
	if len(clause.vars) == 0 {
		removeClause(clause, bve)
	}

	//fmt.Printf("Strengthen result: %s\n", clause.String(bve))
}

func removeClause(clause *Clause, bve *BVEContext) {
	//fmt.Printf("Remove clause: %s\n", clause.String(bve))
	clause.isDeleted = true
	delete(bve.clauses, clause)
	for v := range bve.occur {
		delete(bve.occur[v], clause)
	}

	for v := range clause.vars {
		bve.touched[v] = struct{}{}
	}
}

// Remove any clause subsumed by the first argument
func subsume(clause *Clause, bve *BVEContext) {
	clausesToRemove := findSubsumed(clause, bve)
	for _, c := range clausesToRemove {
		removeClause(c, bve)
	}
}

func removeVariable(varID int64, bve *BVEContext) {
	for c := range bve.clauses {
		delete(c.vars, varID)
		if len(c.vars) == 0 {
			removeClause(c, bve)
		}
	}
	delete(bve.occur, varID)
}

func removeDangling(bve *BVEContext) bool {
	detectedChange := false
	for varID, occ := range bve.occur {
		if occNeg, ok := bve.occur[-varID]; ok {
			if len(occ) == 1 && len(occNeg) == 0 {
				removeVariable(varID, bve)
				detectedChange = true
			}
		} else {
			if len(occ) == 1 {
				removeVariable(varID, bve)
				detectedChange = true
			}
		}
	}
	return detectedChange
}

func removeDuplicates(bve *BVEContext) bool {
	toRemove := []*Clause{}
	for v1 := range bve.clauses {
		if !v1.isDeleted {
			for v2 := range bve.clauses {
				if !v2.isDeleted {
					if v1 != v2 && len(v1.vars) < 10 && len(v2.vars) < 10 {
						if !notEqual(v1, v2, bve) {
							v1.isDeleted = true
							toRemove = append(toRemove, v1)
						}
					}
				}
			}
		}
	}
	for _, c := range toRemove {
		removeClause(c, bve)
	}
	return len(toRemove) > 0
}

func maybeEliminate(varID int64, bve *BVEContext) {
	if len(bve.occur[varID]) > 10 || len(bve.occur[-varID]) > 10 {
		return // Heuristic cut-off
	}
	maybeClauseDistribute(varID, bve)
}

func propagateToplevel(bve *BVEContext) {
	// TODO: Implement
}

/*
 * Eliminates x by clause distribution if the result has fewer clauses than the original
 * (after removing trivially satisfied clauses)
 */
func maybeClauseDistribute(varID int64, bve *BVEContext) {
	// TODO: Implement
}

func cleanup(bve *BVEContext) {
	for {
		r1 := removeDangling(bve)
		r2 := removeDuplicates(bve)
		if !(r1 && r2) {
			break
		}
	}
}

func simplify(bve *BVEContext) {

	/*
	 * Set of variables
	 * A variable is added to this set if it occurs in a clause being added, removed, or strengthened. Initially all variables are "touched"
	 */
	bve.touched = map[int64]struct{}{}
	for clause := range bve.clauses {
		for v := range clause.vars {
			bve.touched[v] = struct{}{}
		}
	}

	/*
	 * Set of clauses
	 * When a clause is added to the SAT problem (e.g. by variable elimination), it is also added to this set.
	 * Initially all clauses are considered "added".
	 */
	bve.added = map[*Clause]struct{}{}
	for clause := range bve.clauses {
		bve.added[clause] = struct{}{}
	}

	/*
	 * Set of clauses
	 * When a clause is strengthened (one literal is removed, either by self-subsumption or toplevel propagation)
	 * it is added to this set. Initially the set is empty.
	 */
	bve.strenghtened = map[*Clause]struct{}{}

	for {
		// Subsumption

		//fmt.Printf("Iterate added %d\n", len(bve.added))

		S0 := getAddedClauseCandidates(&bve.added,  true, bve)
		for {
			//fmt.Printf("Iterate strenghtened\n")

			S1 := getAddedClauseCandidates(&bve.added, false, bve)
			for a := range bve.added {
				S1[a] = struct{}{}
			}
			for s := range bve.strenghtened {
				S1[s] = struct{}{}
			}
			// Clear Added and Strengthened
			bve.added = map[*Clause]struct{}{}
			bve.strenghtened = map[*Clause]struct{}{}

			// Loop
			for c := range S1 {
				selfSubsume(c, bve)
			}
			// May strengthen/remove clauses
			propagateToplevel(bve)

			if len(bve.strenghtened) == 0 {
				break
			}
		}

		//fmt.Printf("Subsuming S0\n")
		for c := range S0 {
			if !c.isDeleted {
				subsume(c, bve)
			}
		}

		// Variable elimination

		//fmt.Printf("Variable elimination loop\n")
		for {
			//fmt.Printf("Eliminate variables\n")
			S := bve.touched
			bve.touched = map[int64]struct{}{}
			for x := range S {
				maybeEliminate(x, bve)
			}
			if len(bve.touched) == 0 {
				break
			}
		}
		cleanup(bve)

		if len(bve.added) == 0 {
			break
		}
	}

}

func negateClauseVar(clause *Clause, varIDToNegate int64) *Clause {
	newVars := map[int64]struct{}{}
	for varID := range clause.vars {
		if varID == varIDToNegate || varID == -varIDToNegate {
			newVars[-varID] = struct{}{}
		} else {
			newVars[varID] = struct{}{}
		}
	}
	return &Clause{
		hash:  clause.hash,
		vars:  newVars,
	}
}

func selfSubsume(clause *Clause, bve *BVEContext) {
	for v := range clause.vars {
		subsumedBy := findSubsumed(negateClauseVar(clause, v), bve)
		for _, cPrim := range subsumedBy {
			//fmt.Printf("<%s> subsumes %s by %s\n", clause.String(bve), cPrim.String(bve), bve.vars.Reverse(v))
			strenghten(cPrim, -v, bve)
		}
	}
}

func VariableElimination(formula *sat_solver.SATFormula) (error, *sat_solver.SATFormula) {
	f := formula.Formula()
	hashVal := int64(0)

	bve := BVEContext{
		clauses:   map[*Clause]struct{}{},
		occur:     map[int64]map[*Clause]struct{}{},
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

		c.hash = hashVal
		bve.clauses[c] = struct{}{}
	}

	simplify(&bve)

	newFormula := sat_solver.CNFFormula{
		Variables: make([][]int64, len(bve.clauses)),
	}
	i := 0
	for c := range bve.clauses {
		newClause := make([]int64, len(c.vars))
		j := 0
		for v := range c.vars {
			newClause[j] = v
			j++
		}
		newFormula.Variables[i] = newClause
		i++
	}

	return nil, sat_solver.NewSATFormula(&newFormula, bve.vars)
}