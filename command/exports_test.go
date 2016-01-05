package command

import "testing"

type testcase struct {
	in   *exportData
	want bool
}

func innerTest(t *testing.T, cases []testcase) {
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 0, // must be > 0
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
			},
			want: false,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     0, // must be > 0
				from:      0,
				timeFmt:   "msec",
				output:    "json",
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

// TestTimeRangeExportDataIsValid tests if the validity checks for
// time ranges are working.
func TestTimeRangeExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				timeFmt:   "msec",
				output:    "json",
				from:      0,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				timeFmt:   "msec",
				output:    "json",
				to:        10,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				timeFmt:   "msec",
				output:    "json",
				from:      0,
				to:        10,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				timeFmt:   "msec",
				output:    "json",
				from:      0,
				to:        0,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				timeFmt:   "msec",
				output:    "json",
				from:      10,
				to:        0,
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

// TestValRangeExportDataIsValid tests if the validity checks for
// value ranges are working.
func TestValRangeExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId:   1,
				limit:       1,
				from:        0,
				timeFmt:     "msec",
				output:      "json",
				greaterThan: 0,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				lessThan:  10,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId:   1,
				limit:       1,
				from:        0,
				timeFmt:     "msec",
				output:      "json",
				greaterThan: 0,
				lessThan:    10,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId:   1,
				limit:       1,
				from:        0,
				timeFmt:     "msec",
				output:      "json",
				greaterThan: 0,
				lessThan:    0,
			},
			want: true,
		},
		{
			in: &exportData{
				projectId:   1,
				limit:       1,
				from:        0,
				timeFmt:     "msec",
				output:      "json",
				greaterThan: 10,
				lessThan:    0,
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

func TestEqualExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				equal:     "0",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				equal:     "-10",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				equal:     "10",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				equal:     "not a number",
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

func TestOpExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "sum",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "count",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "min",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "max",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "mean",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "not an op",
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

func TestGroupExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "sum",
				groupBy:   "1h",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				operator:  "count",
				groupBy:   "12q", // q not a valid unit
			},
			want: false,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
				groupBy:   "1h", // no operator
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

func TestTimeFmtExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "sec",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "usec",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "timeval",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "not valid",
				output:    "json",
			},
			want: false,
		},
	}

	innerTest(t, cases)
}

func TestOutputExportDataIsValid(t *testing.T) {
	cases := []testcase{
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "json",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "csv",
			},
			want: true,
		},
		{
			in: &exportData{
				projectId: 1,
				limit:     1,
				from:      0,
				timeFmt:   "msec",
				output:    "not valid",
			},
			want: false,
		},
	}

	innerTest(t, cases)
}
