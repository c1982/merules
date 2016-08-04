package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"

	"github.com/jhillyerd/go.enmime"
)

type Rules struct {
	config meConfig
}

func (r *Rules) ApplyRules(messageFile string) {

	if !isFileExists(messageFile) {
		log.Println("file not found:", messageFile)
		return
	}

	e, err := ReadEmail(messageFile)

	if err != nil {
		log.Println("Read Error: ", err)
		panic(err)
	}

	body, err := enmime.ParseMIMEBody(e)

	if err != nil {
		log.Println("Perse MIME Error: ", err)
		panic(err)
	}

	if isIt, msg := r.hasBlockedExtensions(body); isIt {
		r.applyInternalRules(e, messageFile, msg)
		return
	}

	if isIt, msg := r.containsMalwareDomain(body); isIt {
		r.applyInternalRules(e, messageFile, msg)
		return
	}

	if isIt, msg := r.hasPasswordProtectionZipFile(body); isIt {
		r.applyInternalRules(e, messageFile, msg)
		return
	}
}

func (r *Rules) applyInternalRules(e *mail.Message, messageFile string, msg string) {

	subject := decodeRFC2047(e.Header.Get("Subject"))
	msg = strings.Replace(msg, "%2", subject, -1)

	if r.config.DeleteDetectedMail {
		deleteFile(messageFile)
		return
	}

	if r.config.SendReportRecipient {
		r.sendMessageToRecipient(e, messageFile, msg)
	}

	if r.config.SendReportSender {
		err := InjectEmailToOutgoing(e, messageFile, msg)
		if err != nil {
			log.Println("Sender report cannot be send:", err)
		}
	}
}

func (r *Rules) hasPasswordProtectionZipFile(body *enmime.MIMEBody) (bool, string) {
	result := false
	resultMsg := r.config.BlockPassZipMsg

	if len(body.Attachments) == 0 {
		log.Println("Cannot detect password protection zip.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := getSizebyKB(body.Attachments[i].Content())

		if size <= r.config.MaxScanSizeKB {
			if strings.HasSuffix(fileName, ".zip") {
				content := body.Attachments[i].Content()
				result = r.isPasswordProtected(fileName, content)
				log.Println("File is protected with password:", fileName)
				resultMsg = strings.Replace(resultMsg, "%1", fileName, -1)
				if result {
					break
				}
			}
		}
	}

	return result, resultMsg
}

func (r *Rules) hasBlockedExtensions(body *enmime.MIMEBody) (bool, string) {

	result := false
	resultMsg := r.config.BlockExtensionsMsg
	if len(body.Attachments) == 0 {
		log.Println("Email does not have a attachment.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= r.config.MaxScanSizeKB {
			result = r.hasSuffixBlocked(fileName)
			if result {
				log.Println("Attachment blocked:", fileName)
				resultMsg = strings.Replace(resultMsg, "%1", fileName, -1)
				break
			}
		}
	}

	return result, resultMsg
}

func (r *Rules) hasSuffixBlocked(name string) bool {
	result := false

	for i := 0; i < len(r.config.BlockExtensions); i++ {
		if strings.HasSuffix(name, r.config.BlockExtensions[i]) {
			result = true
			break
		}
	}

	return result
}

func (r *Rules) containsMalwareDomain(body *enmime.MIMEBody) (bool, string) {

	result := false
	resultMsg := r.config.ScanMalwareDomainMsg

	blist := r.getBlackListDomainsFromConfig()

	if !r.config.ScanMalwareDomain {
		log.Println("Malware scan disabled.")
		return result, resultMsg
	}

	if len(blist) == 0 {
		log.Println("Black list domain not found.")
		return result, resultMsg
	}

	if result, resultMsg = r.isContainsBody(body.HTML, blist); result {
		return result, resultMsg
	}

	result, resultMsg = r.isContainsBody(body.Text, blist)

	if result {
		return result, resultMsg
	}

	log.Println("Malware domain not found.")
	return result, resultMsg
}

func (r *Rules) isContainsBody(body string, blacklist []string) (bool, string) {
	result := false
	resultMsg := r.config.ScanMalwareDomainMsg

	for i := 0; i < len(blacklist); i++ {
		result = strings.Contains(body, blacklist[i])
		if result {
			log.Println("Contains malware domain in email body:", blacklist[i])
			resultMsg = strings.Replace(resultMsg, "%1", formatDomain(blacklist[i]), -1)
			break
		}
	}

	return result, resultMsg
}

func (r *Rules) getBlackListDomainsFromConfig() []string {

	var lines []string
	blacklistFile := fmt.Sprintf("%s\\blacklist.config", currentPath)
	file, err := os.Open(blacklistFile)

	if err != nil {
		log.Println("Blacklist config cannot be opened:", err)
		return lines
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func (r *Rules) isPasswordProtected(fileName string, content []byte) bool {
	tmpFolder := fmt.Sprintf("%v\\.tmp", r.config.MePath)

	if !isFolderExists(tmpFolder) {
		log.Println("Folder not found:", tmpFolder)
		err := createFolder(tmpFolder)

		if err != nil {
			log.Println("Folder cannot be created. ", tmpFolder, err)
			panic(err)
		}

		log.Println("Folder created:", tmpFolder)
	}

	tmpfile := fmt.Sprintf("%v\\%v", tmpFolder, fileName)

	err := saveFile(tmpfile, content)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	log.Println("File saved:", tmpfile)

	result, err := isPasswordProtected(tmpfile)

	if err != nil {
		panic(err)
	}

	if result {
		log.Println("Yes file is password protected.")
	}

	deleteFile(tmpfile)
	log.Println("File deleted:", tmpfile)

	return result
}

func (r *Rules) sendMessageToRecipient(m *mail.Message, messageFile string, message string) {
	content := ChangeEmailBodyToMessage(m, message)

	err := saveFile(messageFile, []byte(content))

	if err != nil {
		panic(err)
	}
}
