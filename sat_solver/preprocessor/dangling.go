package preprocessor

func (opt *SimpleOptimizer) RemoveDanglingVariables() {
	for opt.tryRemoveDanglingVariables() {}
}

func (opt *SimpleOptimizer) tryRemoveDanglingVariables() bool {
	varsToRemove := map[int64]struct{}{}
	for v, occurs := range opt.occur {
		if len(occurs) > 0 && len(opt.occur[-v]) == 0 {
			// Negated variable does not occur anywhere
			varsToRemove[v] = struct{}{}
		}
	}

	if len(varsToRemove) == 0 {
		return false
	}

	// Remove all unwanted clauses
	for v := range varsToRemove {
		for c := range opt.occur[v] {
			opt.removeClause(c)
		}
	}

	opt.validateState()
	return true
}