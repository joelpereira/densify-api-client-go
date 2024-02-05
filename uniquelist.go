package densify

import (
	"fmt"
	"strings"
)

// a unique list of string values
type UniqueList struct {
	strList map[string]bool
}

// initialize the unique list (map[string]string object)
func (l *UniqueList) Initialize() {
	l.strList = make(map[string]bool)
}

// add a string and ensure it's not already in the list/map
func (l *UniqueList) Add(str string) {
	l.strList[str] = true
}

// output the values as a comma separated value list (for useful error messages only)
func (l *UniqueList) CsvStr() string {
	s := ""
	first := true
	pre := ""
	for key, _ := range l.strList {
		s = fmt.Sprintf("%s%s%s", s, pre, key)
		if first {
			pre = ", "
			first = false
		}
	}
	return s
}

// this adds a new line to the csv output for easier viewing of values for end users (not for machine processing)
func (l *UniqueList) CsvStrWithNewLine() string {
	return strings.ReplaceAll(l.CsvStr(), ", ", ",\n")
}

// func (l *UniqueList) Length() int {
// 	return len(l.strList)
// }

// func (l *UniqueList) List() *[]string {
// 	var ls []string
// 	for _, value := range l.strList {
// 		ls = append(ls, value)
// 	}
// 	return &ls
// }
