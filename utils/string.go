package utils

import (
	"regexp"
	"strings"
)

// Split split strings by word
func Split(str string) []string {
	re := regexp.MustCompile(`(?:^[a-z]|[A-Z]+)[a-z0-9]*`)
	return re.FindAllString(str, 1000)
}

// CamelCase convert a string to camel case
func CamelCase(s string) string {
	ss := Split(s)
	ss[0] = strings.ToLower(ss[0])
	return strings.Join(ss, "")
}

// SnakeCase convert a string to snake case
func SnakeCase(s string) string {
	ss := Split(s)
	result := strings.Join(ss, "_")
	return strings.ToUpper(result)
}
