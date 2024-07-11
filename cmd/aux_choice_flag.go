package cmd

import (
	"fmt"
	"strings"
)

type tChoice struct {
	Allowed []string // Slice of allowed values
	Value   string   // Defaulted value
}

/*
newChoice give a list of allowed flag parameters, where the second argument is the default

	'allowed' slice of allowed strings
	'defaulted' default string
*/
func newChoice(allowed []string, defaulted string) *tChoice {
	return &tChoice{
		Allowed: allowed,
		Value:   defaulted,
	}
}

/*
Set is called upon flag provisioning, validates its value

	'p' string to be put into the flag
*/
func (a *tChoice) Set(p string) error {

	isIncluded := func(opts []string, val string) bool {
		for _, opt := range opts {
			if val == opt {
				return true
			}
		}
		return false
	}

	if !isIncluded(a.Allowed, p) {
		return fmt.Errorf("%s is not included in {%s}", p, strings.Join(a.Allowed, "|"))
	}

	/* Set the value */

	a.Value = p

	return nil
}

/*
String returns flag value
*/
func (a tChoice) String() string {
	return a.Value
}

/*
Type returns text for help purposes
*/
func (a *tChoice) Type() string {
	return fmt.Sprintf("{%s}", strings.Join(a.Allowed, "|"))
}
