package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/mail"
	"os"

	"github.com/alexmullins/zip"
	mo "github.com/mohamedattahri/mail"
)

func ReadEmail(file string) (*mail.Message, error) {

	f, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(f)

	m, err := mail.ReadMessage(buf)

	if err != nil {
		return nil, err
	}

	return m, err
}

func ReadEmailBody(file string) (string, error) {
	msg, err := ReadEmail(file)

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(msg.Body)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", body), err
}

func PrintEmail(m *mail.Message, body string) string {

	msg := mo.NewMessage()

	for key, values := range m.Header {
		for idx, _ := range values {

			if key == "Content-Type" {
				msg.SetHeader(key, "text/plain")
			} else {
				msg.SetHeader(key, decodeRFC2047(values[idx]))
			}
		}
	}

	body = fmt.Sprintf("%s %s", body, conf.EmailFooter)
	fmt.Fprintf(msg.Body, body)

	return fmt.Sprint(msg)
}

func decodeRFC2047(s string) string {

	// GO 1.5 does not decode headers, but this may change in future releases...
	decoded, err := (&mime.WordDecoder{}).DecodeHeader(s)
	if err != nil || len(decoded) == 0 {
		return s
	}
	return decoded
}

func isPasswordProtected(file string) (bool, error) {
	result := false

	isen, err := zip.OpenReader(file)

	if err != nil {
		return result, err
	}

	defer isen.Close()

	if len(isen.File) > 0 {
		result = isen.File[0].IsEncrypted()
	}

	return result, err
}

func deleteFile(file string) {

	err := os.Remove(file)

	if err != nil {
		log.Println("Delete error: ", err)
	}

}

func isFolderExists(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func createFolder(path string) error {
	return os.Mkdir(path, 0644)
}

func getSizebyKB(content []byte) int {
	return binary.Size(content) / 1024
}
