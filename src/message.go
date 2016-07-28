package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/mail"

	"github.com/alexmullins/zip"
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

func isPasswordProtected(file string) bool {
	isen, _ := zip.OpenReader(file)
	defer isen.Close()

	return isen.File[0].IsEncrypted()
}
