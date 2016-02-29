package command

import "testing"

const testDescInvalidProjectId = "invalid project id (< 1)"

type dataTestCase struct {
	desc string
	in   Data
	want bool
}

func runDataTestCase(t *testing.T, cases []dataTestCase) {
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Errorf("case '%s' failed", c.desc)
		}
	}
}
