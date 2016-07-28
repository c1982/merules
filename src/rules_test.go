package main

import (
	"testing"

	"github.com/alexmullins/zip"
)

var meessageFile = "O:\\Projects\\Go\\src\\merules\\mai\\1CF842E1C83B4267838199F7B5ACB0FF.MAI"

func Test_Attachments(t *testing.T) {

	var r = Rules{}
	r.ApplyRules(meessageFile)
}

func Test_EncryptedZip(t *testing.T) {

	isen, _ := zip.OpenReader("O:\\Projects\\Go\\src\\merules\\mai\\d.zip")

	t.Logf("is Password: %v", isen.File[0].IsEncrypted())

	isen.Close()
}
