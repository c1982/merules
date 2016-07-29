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
	mo "github.com/mohamedattahri/mail"
)

type Rules struct {
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
	resultMsg := conf.BlockPassZipMsg

	if len(body.Attachments) == 0 {
		log.Println("Attachment not found.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= conf.MaxScanSizeKB {
			if strings.HasSuffix(fileName, ".zip") {
				content := body.Attachments[i].Content()
				result = r.isPasswordProtected(fileName, content)
				log.Println("%v is encrypted", fileName)
				resultMsg = strings.Replace(resultMsg, "%1", fmt.Sprintf(" %v is encrypted like malware.", fileName), 0)
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
	resultMsg := conf.BlockExtensionsMsg

	if len(body.Attachments) == 0 {
		log.Println("Attachment extensions not found.")
		return result, resultMsg
	}

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= conf.MaxScanSizeKB {
			if result = r.hasSuffixBlocked(fileName); result {
				log.Println("File extension blocked: %s", fileName)
				resultMsg = strings.Replace(resultMsg, "%1", fmt.Sprintf("Blocked extension: %v.", fileName), 0)
				break
			}
		}
	}

	return result, resultMsg
}

func (r *Rules) hasSuffixBlocked(name string) bool {
	result := false

	for i := 0; i < len(conf.BlockExtensions); i++ {
		if strings.HasSuffix(name, conf.BlockExtensions[i]) {
			result = true
			break
		}
	}

	return result
}

func (r *Rules) containsMalwareDomain(body *enmime.MIMEBody) (bool, string) {

	result := false
	resultMsg := conf.ScanMalwareDomainMsg

	blist := r.getBlackListDomainsFromConfig()

	if !conf.ScanMalwareDomain {
		log.Println("Malware scan disabled.")
		return result, resultMsg
	}

	if len(blist) == 0 {
		log.Println("Black list domains not found.")
		return result, resultMsg
	}
	for i := 0; i < len(blist); i++ {
		result = strings.ContainsAny(body.HTML, blist[i])
		if result {
			log.Println("Contains malware domain in email body: %s", blist[i])
			resultMsg = strings.Replace(resultMsg, "%1", blist[i], 0)
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
	tmpFolder := fmt.Sprintf("%v\\.tmp", conf.MePath)

	if isFolderExists(tmpFolder) {
		err := createFolder(tmpFolder)
		if err != nil {
			log.Println("Folder cannot created. ", tmpFolder, err)
			panic(err)
		}
	}

	tmpfile := fmt.Sprintf("%v\\%v", tmpFolder, fileName)

	err := r.saveAttachmentFile(tmpfile, content)

	if err != nil {
		log.Println()
		panic(err)
	}

	result, err := isPasswordProtected(tmpfile)

	if err != nil {
		panic(err)
	}

	deleteFile(tmpfile)

	return result
}

func (r *Rules) saveAttachmentFile(fileName string, content []byte) error {
	return ioutil.WriteFile(fileName, content, 0644)
}

func (r *Rules) newPlainTextMsg(messageFile string, m *mail.Message, message string) {
	msg := mo.NewMessage()
	msg.SetHeader("Received", m.Header.Get("Received"))
	msg.SetHeader("From", m.Header.Get("From"))
	msg.SetHeader("References", m.Header.Get("References"))
	msg.SetHeader("Date", m.Header.Get("Date"))
	msg.SetMessageID(m.Header.Get("Message-ID"))
	msg.SetInReplyTo(m.Header.Get("In-Reply-To"))
	msg.SetSubject(m.Header.Get("Subject"))
	msg.SetContentType("text/plain")

	content := []byte(fmt.Sprint(msg))

	err := ioutil.WriteFile(messageFile, content, 0644)

	if err != nil {
		panic(err)
	}
}
