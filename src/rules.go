package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
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
		r.newPlainTextMsg(messageFile, e, msg)
		return
	}

	if isIt, msg := r.containsMalwareDomain(body); isIt {
		r.newPlainTextMsg(messageFile, e, msg)
		return
	}

	if isIt, msg := r.hasPasswordProtectionZipFile(body); isIt {
		r.newPlainTextMsg(messageFile, e, msg)
		return
	}
}

func (r *Rules) hasPasswordProtectionZipFile(body *enmime.MIMEBody) (bool, string) {
	result := false
	resultMsg := r.config.BlockPassZipMsg

	if len(body.Attachments) == 0 {
		log.Println("Attachment not found.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := getSizebyKB(body.Attachments[i].Content())

		if size <= r.config.MaxScanSizeKB {
			if strings.HasSuffix(fileName, ".zip") {
				content := body.Attachments[i].Content()
				result = r.isPasswordProtected(fileName, content)
				log.Println("File encrypted:", fileName)
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
		log.Println("Attachment extensions not found.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= r.config.MaxScanSizeKB {
			result = r.hasSuffixBlocked(fileName)
			if result {
				log.Printf("File extension blocked: %s", fileName)
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
		log.Println("Black list domains not found.")
		return result, resultMsg
	}

	if result, resultMsg = r.isContainsBody(body.HTML, blist); result {
		log.Println("Malware found in HTML part: ", resultMsg)
		return result, resultMsg
	}

	result, resultMsg = r.isContainsBody(body.Text, blist)

	if result {
		log.Println("Malware found in TEXT part: ", resultMsg)
	}

	return result, resultMsg
}

func (r *Rules) isContainsBody(body string, blacklist []string) (bool, string) {
	result := false
	resultMsg := r.config.ScanMalwareDomainMsg

	for i := 0; i < len(blacklist); i++ {
		result = strings.Contains(body, blacklist[i])
		if result {
			log.Println("Contains malware domain in email body:", blacklist[i])
			resultMsg = strings.Replace(resultMsg, "%1", blacklist[i], -1)
			break
		}
	}

	return result, resultMsg
}

func (r *Rules) getBlackListDomainsFromConfig() []string {

	var lines []string

	file, err := os.Open(".\\blacklist.config")

	if err != nil {
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
			log.Println("Folder cannot created. ", tmpFolder, err)
			panic(err)
		}

		log.Println("Folder created:", tmpFolder)
	}

	tmpfile := fmt.Sprintf("%v\\%v", tmpFolder, fileName)

	err := r.saveAttachmentFile(tmpfile, content)

	if err != nil {
		log.Println()
		panic(err)
	}

	log.Println("File saved:", tmpfile)

	result, err := isPasswordProtected(tmpfile)

	if err != nil {
		panic(err)
	}

	deleteFile(tmpfile)
	log.Println("File deleted:", tmpfile)

	return result
}

func (r *Rules) saveAttachmentFile(fileName string, content []byte) error {
	return ioutil.WriteFile(fileName, content, 0644)
}

func (r *Rules) newPlainTextMsg(messageFile string, m *mail.Message, message string) {
	content := PrintEmail(m, message)

	err := ioutil.WriteFile(messageFile, []byte(content), 0644)

	if err != nil {
		panic(err)
	}
}
