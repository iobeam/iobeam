package command

import (
	"regexp"
)

const VALUE_DOUBLE_REGEXP = "^[-+]?[0-9]+.[0-9]+$"
const VALUE_LONG_REGEXP = "^[-+]?[0-9]+$"
const VALUE_STRING_REGEXP = "^\".*\"$"

func IsValidTypeString(s string) bool {
	switch s {
	case
		"DOUBLE",
		"LONG",
		"STRING":
		return true
	}
	return false
}

func IsValidDouble(value string) bool {
	matched, err := regexp.MatchString(VALUE_DOUBLE_REGEXP, value)
	if err == nil && matched {
		return true
	}
	return false
}

func IsValidLong(value string) bool {
	matched, err := regexp.MatchString(VALUE_LONG_REGEXP, value)
	if err == nil && matched {
		return true
	}
	return false
}

func IsValidString(value string) bool {
	matched, err := regexp.MatchString(VALUE_STRING_REGEXP, value)
	if err == nil && matched {
		return true
	}
	return false
}

func IsValidType(value string) bool {
	return IsValidString(value) || IsValidLong(value) || IsValidDouble(value)
}
