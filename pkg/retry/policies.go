package retry

import "fmt"

type Policy int

const (
	PolicyLinear Policy = iota
	PolicyBackoff
	PolicyInfinite
)

// Validate validates the retry policy
func (r Policy) Validate() error {
	switch r {
	case PolicyLinear, PolicyBackoff, PolicyInfinite:
		return nil
	default:
		return fmt.Errorf("invalid retry policy")
	}
}
