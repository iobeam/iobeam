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
