package sat_solver

import "fmt"

type SATVariableMapping struct {
	names map[string]int64
	reverse map[int64]string
	uniqueID int64
	freshVarNameID uint64
}

func NewSATVariableMapping() *SATVariableMapping {
	return &SATVariableMapping{
		names:    map[string]int64{},
		reverse:  map[int64]string{},
		uniqueID: 2,
		freshVarNameID: 1,
	}
}

func (vars *SATVariableMapping) Reverse(id int64) string {
	if id < 0 {
		return fmt.Sprintf("-%s", trimVarQuotes(vars.reverse[-id]))
	}
	return trimVarQuotes(vars.reverse[id])
}

func (vars *SATVariableMapping) Fresh() (string, int64) {
	newVarNameID := vars.freshVarNameID
	newID := vars.uniqueID
	name := fmt.Sprintf("[%d]", newVarNameID)
	vars.uniqueID++
	vars.freshVarNameID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return name, newID
}

func (vars *SATVariableMapping) Get(name string) int64 {
	if id, ok := vars.names[name]; ok {
		return id
	}
	newID := vars.uniqueID
	vars.uniqueID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return newID
}
