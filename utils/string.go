package utils

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jinzhu/inflection"
)

// Split split strings by word
func Split(str string) []string {
	strs := strings.Split(str, "_")
	for i, s := range strs {
		strs[i] = upperFirst(s)
	}
	re := regexp.MustCompile(`(?:^[a-z]|[A-Z]+)[a-z0-9]*`)
	return re.FindAllString(strings.Join(strs, ""), 1000)
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// CamelCase convert a string to camel case
func CamelCase(s string) string {
	ss := Split(s)
	for i := range ss {
		ss[i] = strings.ToLower(ss[i])
		return strings.Join(ss, "")
	}
	return ""
}

// UpperSnakeCase convert a string to snake case
func UpperSnakeCase(s string) string {
	ss := Split(s)
	result := strings.Join(ss, "_")
	return strings.ToUpper(result)
}

// InterfaceCase convert a string to interface name start with I
func InterfaceCase(s string) string {
	ss := Split(s)
	return "I" + strings.Join(ss, "")
}

// PluralCase return the plural of a word
func PluralCase(word string) string {
	return inflection.Plural(word)
}

// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//   "" =>                     [""]
//   "lowercase" =>            ["lowercase"]
//   "Class" =>                ["Class"]
//   "MyClass" =>              ["My", "Class"]
//   "MyC" =>                  ["My", "C"]
//   "HTML" =>                 ["HTML"]
//   "PDFLoader" =>            ["PDF", "Loader"]
//   "AString" =>              ["A", "String"]
//   "SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//   "vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//   "GL11Version" =>          ["GL", "11", "Version"]
//   "99Bottles" =>            ["99", "Bottles"]
//   "May5" =>                 ["May", "5"]
//   "BFG9000" =>              ["BFG", "9000"]
//   "BöseÜberraschung" =>     ["Böse", "Überraschung"]
//   "Two  spaces" =>          ["Two", "  ", "spaces"]
//   "BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1) If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2) Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3) Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4) Iterate through array of split strings, and if a given string
//     is upper case:
//       if subsequent string is lower case:
//         move last character of upper case string to beginning of
//         lower case string
// func Split(src string) (entries []string) {
// 	// don't split invalid utf8
// 	if !utf8.ValidString(src) {
// 		return []string{src}
// 	}
// 	entries = []string{}
// 	var runes [][]rune
// 	lastClass := 0
// 	class := 0
// 	// split into fields based on class of unicode character
// 	for _, r := range src {
// 		switch true {
// 		case unicode.IsLower(r):
// 			class = 1
// 		case unicode.IsUpper(r):
// 			class = 2
// 		case unicode.IsDigit(r):
// 			class = 3
// 		default:
// 			class = 4
// 		}
// 		if class == lastClass {
// 			runes[len(runes)-1] = append(runes[len(runes)-1], r)
// 		} else {
// 			runes = append(runes, []rune{r})
// 		}
// 		lastClass = class
// 	}
// 	// handle upper case -> lower case sequences, e.g.
// 	// "PDFL", "oader" -> "PDF", "Loader"
// 	for i := 0; i < len(runes)-1; i++ {
// 		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
// 			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
// 			runes[i] = runes[i][:len(runes[i])-1]
// 		}
// 	}
// 	// construct []string from results
// 	for _, s := range runes {
// 		if len(s) > 0 {
// 			entries = append(entries, string(s))
// 		}
// 	}
// 	return
// }
