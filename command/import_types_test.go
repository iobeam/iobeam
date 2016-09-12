package command

import (
	"testing"
)

func TestIsValidDouble(t *testing.T) {

	validDoubles := [...]string{"0.43", "10.2", "-2.3", "3.0"}
	invalidDoubles := [...]string{"", "a", "1", "0.2.2", ".2"}

	for _, validDouble := range validDoubles {
		if !IsValidDouble(validDouble) {
			t.Fatalf("IsValidDouble(%s) returned false", validDouble)
		}
	}

	for _, invalidDouble := range invalidDoubles {
		if IsValidDouble(invalidDouble) {
			t.Fatalf("IsValidDouble(%s) returned true", invalidDouble)
		}
	}
}

func TestIsValidString(t *testing.T) {

	validStrings := [...]string{"\"\"", "\"foo\""}
	invalidStrings := [...]string{"", "foo"}

	for _, validString := range validStrings {
		if !IsValidString(validString) {
			t.Fatalf("IsValidString(%s) returned false", validString)
		}
	}

	for _, invalidString := range invalidStrings {
		if IsValidString(invalidString) {
			t.Fatalf("IsValidString(%s) returned true", invalidString)
		}
	}
}

func TestIsValidLong(t *testing.T) {

	validLongs := [...]string{"102", "0", "-2"}
	invalidLongs := [...]string{"", "foo", "0.2"}

	for _, validLong := range validLongs {
		if !IsValidLong(validLong) {
			t.Fatalf("IsValidLong(%s) returned false", validLong)
		}
	}

	for _, invalidLong := range invalidLongs {
		if IsValidLong(invalidLong) {
			t.Fatalf("IsValidLong(%s) returned true", invalidLong)
		}
	}
}

func TestIsValidBoolean(t *testing.T) {

	validBooleans := [...]string{"true", "false", "True", "FalSe"}
	invalidBooleans := [...]string{"flase", "Truee", "ffalsee"}

	for _, validBoolean := range validBooleans {
		if !IsValidBoolean(validBoolean) {
			t.Fatalf("IsValidBoolean(%s) returned false", validBoolean)
		}
	}

	for _, invalidBoolean := range invalidBooleans {
		if IsValidBoolean(invalidBoolean) {
			t.Fatalf("IsValidBoolean(%s) returned true", invalidBoolean)
		}
	}
}
