package main

import (
	"bufio"
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

func ReadEmailHeaders(e *mail.Message) string {

	var b bytes.Buffer

	b.WriteString("\r\n\r\n\r\nDiagnostic information for administrators:\r\n\r\n")

	for key, values := range e.Header {
		for idx, _ := range values {

			b.WriteString(fmt.Sprintf("%s=%s\r\n", key, decodeRFC2047(values[idx])))
		}
	}

	return b.String()
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

	msg.SetSubject("Received Failure")

	debugInfo := ReadEmailHeaders(m)

	body = fmt.Sprintf("%s %s %s", body, debugInfo, conf.EmailFooter)
	fmt.Fprintf(msg.Body, body)

	return fmt.Sprint(msg)
}

func InjectEmailToOutgoing(e *mail.Message, messageFile string, msg string) error {

	sender := conf.SenderEmail
	messageId := getFileNameOfPath(messageFile)

	commandContent := createCommandFileContent(e, sender, messageId)
	mailContent := createNDRContent(e, messageFile, msg)

	outgoingCommand := fmt.Sprintf("%v\\Queues\\SMTP\\Outgoing\\%v", conf.MePath, messageId)
	outgoingMessage := fmt.Sprintf("%v\\Queues\\SMTP\\Outgoing\\Messages\\%v", conf.MePath, messageId)

	err := saveFile(outgoingCommand, []byte(commandContent))

	if err != nil {
		return err
	}

	err = saveFile(outgoingMessage, []byte(mailContent))

	return err
}

func createNDRContent(e *mail.Message, messageFile string, msg string) string {

	to, err := mo.ParseAddress(e.Header.Get("From"))

	if err != nil {
		log.Println("From cannot be parsed", err)
		return ""
	}

	messageId := decodeRFC2047(e.Header.Get("Message-ID"))

	m := mo.NewMessage()
	m.SetFrom(&mo.Address{"Postmaster", conf.SenderEmail})
	m.SetContentType("text/plain")
	m.SetMessageID(messageId)
	m.SetSubject("Delivery Failure")
	m.To().Add(to)

	debugInfo := ReadEmailHeaders(e)

	body := fmt.Sprintf("%s %s %s", msg, debugInfo, conf.EmailFooter)

	fmt.Fprintf(m.Body, body)

	return m.String()
}

func createCommandFileContent(e *mail.Message, sender string, messageId string) string {

	to, err := mo.ParseAddress(e.Header.Get("From"))

	if err != nil {
		log.Println("From cannot be parsed", err)
		return ""
	}

	toDomain := strings.Split(to.Address, "@")[1]
	senderDomain := strings.Split(sender, "@")[1]

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("DomainName=%s\r\n", toDomain))
	b.WriteString("CommandType=NDR\r\n")
	b.WriteString(fmt.Sprintf("Recipients=[SMTP:%s]\r\n", to.Address))
	b.WriteString(fmt.Sprintf("Sender=[SMTP:%s]\r\n", sender))
	b.WriteString("Retries=0\r\n")
	b.WriteString(fmt.Sprintf("MessageID=%s\r\n", messageId))
	b.WriteString(fmt.Sprintf("User=%s\r\n", sender))
	b.WriteString(fmt.Sprintf("Account=%s\r\n", senderDomain))
	b.WriteString("Priority=Normal\r\n")
	b.WriteString("Status=Unsent\r\n")

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

func isPermittedService(services []string, service string) bool {

	result := false

	for _, v := range services {
		if v == service {
			result = true
			break
		}
	}

	return result
}

func deleteFile(path string) {

	err := os.Remove(path)

	if err != nil {
		log.Println("File cannot be deleted:", path, err)
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

func saveFile(fileName string, content []byte) error {
	return ioutil.WriteFile(fileName, content, 0644)
}

func formatDomain(name string) string {
	return strings.Replace(name, ".", "_", -1)
}

func ReadAllLines(fileName string) []string {
	var lines []string

	file, err := os.Open(fileName)

	if err != nil {
		log.Println("File not found:", err)
		return lines
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func GetBlackListDomains() []string {
	blacklistFile := fmt.Sprintf("%s\\blacklist.config", currentPath)

	return ReadAllLines(blacklistFile)
}

func GetWhiteListDomains() []string {
	whiteListFile := fmt.Sprintf("%s\\whitelist.config", currentPath)

	return ReadAllLines(whiteListFile)
}
