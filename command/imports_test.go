package command

import (
	"testing"
)

type importTestcase struct {
	in   *importData
	want bool
}

func innerImportsTest(t *testing.T, cases []importTestcase) {
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Fatalf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsValid(t *testing.T) {

	cases := []importTestcase{
		{
			in: &importData{
				projectId: 42,
				deviceId:  "device_id",
				series:    "series",
				timestamp: 123,
				value:     "30",
			},
			want: true,
		},
		{
			in: &importData{
				projectId: 0,
				deviceId:  "device_id",
				series:    "series",
				timestamp: 123,
				value:     "42",
			},
			want: false,
		},
		{
			in: &importData{
				projectId: 42,
				deviceId:  "",
				series:    "series",
				timestamp: 123,
				value:     "42",
			},
			want: false,
		},
		{
			in: &importData{
				projectId: 42,
				deviceId:  "device_id",
				series:    "",
				timestamp: 123,
				value:     "42",
			},
			want: false,
		},
		{
			in: &importData{
				projectId: 42,
				deviceId:  "device_id",
				series:    "series",
				timestamp: -1,
				value:     "42",
			},
			want: false,
		}, {
			in: &importData{
				projectId: 42,
				deviceId:  "device_id",
				series:    "series",
				timestamp: 123,
				value:     "",
			},
			want: false,
		},
	}

	innerImportsTest(t, cases)

}

func TestCreateDatapoint(t *testing.T) {

	dp, err := createDatapoint(42, "3.0")

	if err != nil {
		t.Fatalf("createDatapoint(42,\"3.0\") failed")
	}

	if dp[1] != 3.0 {
		t.Fatalf("createDatapoint(42,\"2.0\") returned %q", dp)
	}

	dp, err = createDatapoint(42, "20")

	if err != nil {
		t.Fatalf("createDatapoint(42,\"20\") failed")
	}

	if dp[1] != int64(20) {
		t.Fatalf("createDatapoint(42,\"20\") returned %q", dp)
	}

	dp, err = createDatapoint(42, "\"20\"")

	if err != nil {
		t.Fatalf("createDatapoint(42,\"\"20\"\") failed")
	}

	if dp[1] != "20" {
		t.Fatalf("createDatapoint(42,\"\"20\"\") returned %q", dp)
	}

	dp, err = createDatapoint(42, "foo")

	if err == nil {
		t.Fatalf("createDatapoint(42,\"foo\") returned %q when it should have failed", dp)
	}
}
