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
	"strings"

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

func ChangeEmailBodyToMessage(m *mail.Message, body string) string {

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

func InjectEmailToOutgoing(e *mail.Message, messageFile string, msg string) {

	to := e.Header.Get("From")
	sender := conf.SenderEmail
	messageId := getFileNameOfPath(messageFile)

	command := createCommandFileContent(to, sender, messageId)

}

func createCommandFileContent(to string, sender string, messageId string) string {

	toDomain := strings.Split(to, "@")[1]
	senderDomain := strings.Split(sender, "@")[1]

	var b bytes.Buffer

	b.WriteString(fmt.Sprintln("DomainName=%s", toDomain))
	b.WriteString(fmt.Sprintln("CommandType=NDR"))
	b.WriteString(fmt.Sprintln("Recipients=[SMTP:%s]", to))
	b.WriteString(fmt.Sprintln("Sender=[SMTP:%s]", sender))
	b.WriteString(fmt.Sprintln("Retries=0"))
	b.WriteString(fmt.Sprintln("MessageID=%s", messageId))
	b.WriteString(fmt.Sprintln("User=%s", sender))
	b.WriteString(fmt.Sprintln("Account=%s", senderDomain))
	b.WriteString(fmt.Sprintln("Priority=Normal"))
	b.WriteString(fmt.Sprintln("Status=Unsent"))

	return b.String()
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

func isFolderExists(path string) bool {
	return isExistsItem(path, true)
}

func isFileExists(path string) bool {
	return isExistsItem(path, false)
}

func isExistsItem(path string, checkdir bool) bool {
	i, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	if checkdir {
		return i.IsDir()
	}

	return true
}

func isPermittedService(service string) bool {

	result := false

	for _, v := range conf.ScanServices {
		if v == service {
			result = true
			break
		}
	}

	return result
}

func deleteFile(path string) {

	err := os.Remove("deleteme.file")

	if err != nil {
		log.Println("File cannot be deleted:", path)
		return
	}

	log.Println("File deleted:", path)
}

func createFolder(path string) error {
	return os.Mkdir(path, 0644)
}

func getSizebyKB(content []byte) int {
	return binary.Size(content) / 1024
}

func getFileNameOfPath(path string) string {
	i := strings.LastIndex(path, "\\")
	return path[i+1:]
}
