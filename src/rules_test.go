package main

import (
	"testing"
)

var meessageFile = "O:\\Projects\\Go\\src\\merules\\mai\\1CF842E1C83B4267838199F7B5ACB0FF.MAI"

func Test_Attachments(t *testing.T) {

	var r = Rules{}
	r.ApplyRules(meessageFile)
}
