package preprocessor

import "fmt"

/*
 * Eliminates x by clause distribution if the result has fewer clauses than the original
 * (after removing trivially satisfied clauses)
 */
func (opt *SimpleOptimizer) maybeClauseDistribute(varID int64) {
	// TODO: Implement
}

func (opt *SimpleOptimizer) tryDistributeClauses() bool {
	for varID, varOccur := range opt.occur {
		for clause := range varOccur {
			len1 := len(clause.vars)
			if !clause.isDeleted && len1 > 0 {
				for negClause := range opt.occur[-varID] {
					len2 := len(negClause.vars)
					if clause != negClause && !negClause.isDeleted && len2 > 0 {
						fmt.Printf("Combined clauses %s and %s ", clause.String(opt), negClause.String(opt))
						opt.removeClause(negClause)
						if len1 == 1 && len1 + len2 - 2 != 1 {
							delete(opt.singular, clause)
						}
						if len1 != 1 && len1 + len2 - 2 == 1 {
							opt.singular[clause] = struct{}{}
						}
						if len1 == 2 && len1 + len2 - 2 != 2 {
							// Cleanup occurBi lists
							for v, _ := range clause.vars {
								delete(opt.occurBi[v], clause)
							}
						}

						// Update occur lists and add variables to the clause
						delete(clause.vars, varID)
						delete(opt.occur[varID], clause)
						delete(opt.occurBi[varID], clause)

						isTautology := false
						for v := range negClause.vars {
							if v != -varID {
								opt.occur[v][clause] = struct{}{}
								clause.vars[v] = struct{}{}
								if _, ok := opt.occur[-v][clause]; ok {
									isTautology = true
									break
								}
							}
						}

						if len1 != 2 && len1+len2-2 == 2 {
							// Cleanup occurBi lists
							for v := range clause.vars {
								opt.occurBi[v][clause] = struct{}{}
							}
						}

						if isTautology {
							opt.removeClause(clause)
						} else {
							clause.Rehash()
						}
						opt.validateState()

						fmt.Printf("into %s\n", clause.String(opt))
						return true
					}
				}
			}
		}
	}
	return false
}

func (opt *SimpleOptimizer) DistributeClauses() bool {
	for opt.tryDistributeClauses() {}
	return false
}