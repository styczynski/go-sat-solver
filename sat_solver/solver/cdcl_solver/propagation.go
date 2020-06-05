package cdcl_solver

import (
	"strings"

	"github.com/go-sat-solver/sat_solver"
)

type TWL []*TWLRecord

/**
 * Print human-readable representation of a TWL records set.
 */
func (twl TWL) DebugString() string {
	ret := []string{}
	for _, record := range twl {
		ret = append(ret, record.DebugString())
	}
	return strings.Join(ret,",")
}

/*
 * Performs unit propagation using TWLRecord data structure.
 * This data structure is described here:
 *   http://people.mpi-inf.mpg.de/~mfleury/sat_twl.pdf
 *
 * For code reference please see (Minisat solver):
 *   https://github.com/niklasso/minisat/blob/master/minisat/core/Solver.cc#L506
 *
 * The unit propagation use benefit of the Watched Literals algorithm.
 * The key idea is to maintain list of "watched" literals in each clause such that the following invariant holds:
 *
 *    Watched literal is iff the other watched literal is true and all the unwatched literals are false.
 *
 * [Direct citation from the linked work "A Verified SAT Solver with Watched Literals Using Imperative HOL"]:
 *
 *   For each literal L, the clauses that contain a watched L are
 *   chained together in a list, called a watch list. When a literal L
 *   becomes true, the solver needs to iterate only through the
 *   watch list for −L to find candidates for propagation or conflict. For each candidate clause, there are four possibilities:
 *     1. If the other watched literal is true, do nothing.
 *     2. If one of the unwatched literals L' is not false, restore the invariant by updating the clause so that it watches L' instead of −L.
 *     3. Otherwise, consider the other watched literal L' in the clause:
 *       3.1. If it is not set, propagate L'
 *       3.2. Otherwise, L' is false, and we have found a conflict
 *
 * This code follows those checks.
 */
func (solver *CDCLSolver) performUnitPropagation() sat_solver.CNFClause {
	assignmentTraceLength := len(solver.assignmentTrace)

	/*
	 * We go trough all of the trace and perform unit propagation.
	 * We use currentTraceCheckIndex instead of local variable, because when we jump back on conflict we don't want to
	 * waste time on checking previous assignments that didn't change.
	 *
	 * The currentTraceCheckIndex is similar to qhead variable in the Minisat code.
	 */
	for solver.currentTraceCheckIndex < assignmentTraceLength {
		assignmentTraceLength = len(solver.assignmentTrace)
		// Get the next literal assigned in the assignmentTrace
		p := solver.assignmentTrace[solver.currentTraceCheckIndex]
		solver.currentTraceCheckIndex++

		/**
		 * We will write new watched literals into separate slice
		 */
		watchedLiteralsForVar := solver.watchedLiterals[p]
		newWatchedLiterals := make([]*TWLRecord, 0, len(watchedLiteralsForVar))

		/*
		 * Go through all the watched literals inside TWLRecord data structure and check if they're affected by the assignment
		 */
		for i := 0; i < len(watchedLiteralsForVar); i++ {
			watchedLiteral := watchedLiteralsForVar[i]

			/*
			 * If the watched literal is true we don't have to check anything, because the clause is true.
			 * This happens due to TWLRecord invariant.
			 */
			if solver.currentLiteralValue(watchedLiteral.Literal).IsTrue() {
				newWatchedLiterals = append(newWatchedLiterals, watchedLiteral)
			} else {

				watchedLiteralClause := watchedLiteral.Clause
				first := watchedLiteralClause[0]

				/*
				 * If the first literal is -L we swap it so the first is always L'.
				 */
				if first == -p {
					watchedLiteralClause[0], watchedLiteralClause[1] = watchedLiteralClause[1], -p
					first = watchedLiteralClause[0]
				}

				//Debug("[TRACE] sat: %s watched literals2: %s)\n", p.DebugString(), TWL(solver.watchedLiterals[p]).DebugString())
				newWatcher := NewTWLRecord(first, watchedLiteral.Clause)

				/*
				 * If the other watched literal is true, do nothing. (1)
				 */
				if solver.currentLiteralValue(first).IsTrue() && first != watchedLiteral.Literal {
					newWatchedLiterals = append(newWatchedLiterals, newWatcher)
					continue
				}

				/*
				 * If one of the unwatched literals L' is not false,
				 * restore the invariant by updating the clause so that it watches L′ instead of −L (2)
				 */
				detectedUnwatchedNonFalse := false
				for j, literal := range watchedLiteralClause[2:] {
					if !solver.currentLiteralValue(literal).IsFalse() {
						// Watch that literal now (swaps with the second element in watched tuple)
						watchedLiteralClause[1], watchedLiteralClause[j+2] = watchedLiteralClause[j+2], -p
						solver.watchedLiterals[-literal] = append(solver.watchedLiterals[-literal], newWatcher)
						detectedUnwatchedNonFalse = true
						break
					}
				}
				if detectedUnwatchedNonFalse {
					continue
				}

				/*
				 * Otherwise, consider the other watched literal L'in the clause: (3)
				 *
				 * Every unwatched value is false so a clause can be:
				 *   - unit clause (so we shall propagate)
				 *   - conflict (so we must return)
				 */
				newWatchedLiterals = append(newWatchedLiterals, newWatcher)

				/*
				 * If L' is false, then we detected a conflict. (3.2)
				 */
				if solver.currentLiteralValue(first).IsFalse() {
					if i != len(newWatchedLiterals)-1 {
						newWatchedLiterals = append(newWatchedLiterals, watchedLiteralsForVar[i+1:]...)
						solver.watchedLiterals[p] = newWatchedLiterals
					} else {
						newWatchedLiterals = append(newWatchedLiterals, watchedLiteralsForVar[len(newWatchedLiterals):]...)
						solver.watchedLiterals[p] = newWatchedLiterals
					}
					return watchedLiteral.Clause
				}

				/**
				 * L' is undefined, so we propagate. (3.1)
				 */
				solver.performLiteralAssertion(first, watchedLiteral.Clause)
			}
		}

		/*
		 * Getting here means that there is no conflict so far, so we set new watched literals and proceed
		 * to the next literal.
		 */
		solver.watchedLiterals[p] = newWatchedLiterals
		assignmentTraceLength = len(solver.assignmentTrace)
	}

	return nil
}

/**
 * Create TWL records for a specified clause.
 * Please note that the given clause must contain at least two literals.
 * For efficiency reasons any length checks are skipped.
 */
func (solver *CDCLSolver) watchClause(clause sat_solver.CNFClause) {
	solver.watchedLiterals[-clause[0]] = append(solver.watchedLiterals[-clause[0]], &TWLRecord{
		Literal: clause[1],
		Clause:  clause,
	})
	solver.watchedLiterals[-clause[1]] = append(solver.watchedLiterals[-clause[1]], &TWLRecord{
		Literal: clause[0],
		Clause:  clause,
	})
}
