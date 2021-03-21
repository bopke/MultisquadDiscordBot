package util

import "testing"

func TestMentionRegex(t *testing.T) {
	str := "<@205745502266851329>"
	if !IsMention(str) {
		t.Error(str, "should be correct mention!")
	}
	str = "<@!205745502266851329>"
	if !IsMention(str) {
		t.Error(str, "should be correct mention!")
	}
	str = "<205745502266851329>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "<@205745502266851329"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "@205745502266851329>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "<@aaa>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "<@>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "<@!>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
	str = "<@!abc>"
	if IsMention(str) {
		t.Error(str, "shouldn't be correct mention!")
	}
}
