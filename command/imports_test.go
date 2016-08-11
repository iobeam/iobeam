package command

import (
	"reflect"
	"testing"
)

type importTestcase struct {
	in   *importData
	want bool
}

func innerImportsTest(t *testing.T, cases []importTestcase) {
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Fatalf("IsValid(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsValid(t *testing.T) {

	cases := []importTestcase{
		{
			in: &importData{
				projectId: 1,
				fields:    "fields",
				timestamp: 123,
				values:    "30",
			},
			want: true,
		},
		{
			in: &importData{
				projectId: 0,
				fields:    "fields",
				timestamp: 123,
				values:    "42",
			},
			want: false,
		},
		{
			in: &importData{
				projectId: 5,
				fields:    "fields",
				timestamp: -1,
				values:    "42",
			},
			want: false,
		}, {
			in: &importData{
				projectId: 6,
				fields:    "fields",
				timestamp: 123,
				values:    "",
			},
			want: false,
		},
	}

	innerImportsTest(t, cases)

}

func TestStrToFields(t *testing.T) {
	s1 := "foo,bar,baz"
	correct1 := []string{"time", "foo", "bar", "baz"}
	fields, addTime := strToFieldNames(s1)
	if !reflect.DeepEqual(fields, correct1) {
		t.Fatalf("Failed to parse %s, got %v", s1, fields)
	}

	if !addTime {
		t.Fatalf("Should return addTime true.")
	}

	s2 := "time,foo,bar,baz"
	correct2 := []string{"time", "foo", "bar", "baz"}
	fields, addTime = strToFieldNames(s2)

	if !reflect.DeepEqual(fields, correct2) {
		t.Fatalf("Failed to parse %s, got %v", s2, fields)
	}

	if addTime {
		t.Fatalf("Should return addTime false.")
	}
}

func TestStrToValues(t *testing.T) {

	s1 := "1,2.0,\"foo\""
	correct1 := []interface{}{int64(1), float64(2.0), "foo"}

	values, _ := strToValues(s1, 3, false)

	if !reflect.DeepEqual(values, correct1) {
		t.Fatalf("Failed to parse %s, got %v", s1, values)
	}

	values, _ = strToValues(s1, 4, true)

	if len(values) != 4 || !reflect.DeepEqual(values[1:4], correct1) {
		t.Fatalf("Failed to parse %s, got %v", s1, values)
	}
}

func TestStrToValue(t *testing.T) {

	dp, err := strToValue("3.0")

	if err != nil {
		t.Fatalf("strToValue(\"3.0\") failed")
	}

	if dp != float64(3.0) {
		t.Fatalf("strToValue(\"2.0\") returned %q", dp)
	}

	dp, err = strToValue("20")

	if err != nil {
		t.Fatalf("strToValue(\"20\") failed")
	}

	if dp != int64(20) {
		t.Fatalf("strToValue(\"20\") returned %q", dp)
	}

	dp, err = strToValue("\"20\"")

	if err != nil {
		t.Fatalf("strToValue(\"\"20\"\") failed")
	}

	if dp != "20" {
		t.Fatalf("strToValue(\"\"20\"\") returned %q", dp)
	}

	dp, err = strToValue("foo")

	if err == nil {
		t.Fatalf("strToValue(\"foo\") returned %q when it should have failed", dp)
	}
}
