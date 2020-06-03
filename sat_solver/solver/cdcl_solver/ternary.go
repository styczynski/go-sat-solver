package cdcl_solver

import "fmt"

/**
 * Boolean type with undefined state.
 */
type Ternary int8

/**
 * Ternary logic states
 */
const (
	TERNARY_TRUE      Ternary = 0
	TERNARY_FALSE             = 1
	TERNARY_UNDEFINED         = 2
)

/**
 * Check if the value is true.
 */
func (t Ternary) IsTrue() bool {
	return t == TERNARY_TRUE
}

/**
 * Check if the value is false.
 */
func (t Ternary) IsFalse() bool {
	return t == TERNARY_FALSE
}

/**
 * Check if the value is undefined.
 */
func (t Ternary) IsUndefined() bool {
	return t == TERNARY_UNDEFINED
}

/**
 * Convert ternary value to boolean.
 * Note:
 *   If the value is undefined then the function will panic.
 */
func TernaryToBool(value Ternary) bool {
	if value == TERNARY_UNDEFINED {
		panic(fmt.Errorf("Tried to convert ternary undefined to bool"))
	}
	return value == TERNARY_TRUE
}

/**
 * Convert bool value to ternary.
 */
func BoolToTernary(value bool) Ternary {
	if value {
		return TERNARY_TRUE
	}
	return TERNARY_FALSE
}