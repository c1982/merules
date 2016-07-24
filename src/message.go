package main

import (
	"bytes"
	"io/ioutil"
	"net/mail"
)

func ReadEmail(file string) (*mail.Message, error) {

	f, _ := ioutil.ReadFile(file)
	buf := bytes.NewBuffer(f)

	return mail.ReadMessage(buf)
}
