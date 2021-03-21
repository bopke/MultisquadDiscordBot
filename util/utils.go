package util

import (
	"regexp"
)

var (
	mentionRegex = regexp.MustCompile("^(<@!?\\d+>)$")
)

func IsMention(str string) bool {
	return mentionRegex.MatchString(str)
}
