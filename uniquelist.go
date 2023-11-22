package densify

import (
	"fmt"
	"strings"
)

type UniqueList struct {
	strList map[string]string
}

func (l *UniqueList) Init() {
	l.strList = make(map[string]string)
}

func (l *UniqueList) Add(str string) {
	l.strList[str] = str
}

func (l *UniqueList) List() *[]string {
	var ls []string
	for _, value := range l.strList {
		ls = append(ls, value)
	}
	return &ls
}

func (l *UniqueList) CsvStr() string {
	s := ""
	first := true
	pre := ""
	for _, value := range l.strList {
		s = fmt.Sprintf("%s%s%s", s, pre, value)
		if first {
			pre = ", "
			first = false
		}
	}
	return s
}

func (l *UniqueList) CsvStrWithNewLine() string {
	return strings.ReplaceAll(l.CsvStr(), ", ", ",\n")
}

func (l *UniqueList) Length() int {
	return len(l.strList)
}
