package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/mail"
)

func ReadEmail(file string) (*mail.Message, error) {

	f, _ := ioutil.ReadFile(file)
	buf := bytes.NewBuffer(f)

	return mail.ReadMessage(buf)
}

func ReadEmailBody(file string) (string, error) {
	msg, err := ReadEmail(file)
	body, err := ioutil.ReadAll(msg.Body)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", body), err
}
