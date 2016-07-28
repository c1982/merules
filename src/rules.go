package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"

	"github.com/jhillyerd/go.enmime"
	mo "github.com/mohamedattahri/mail"
)

type Rules struct {
}

func (r *Rules) ApplyRules(messageFile string) {

	e, _ := ReadEmail(messageFile)
	body, _ := enmime.ParseMIMEBody(e)

	r.newPlainTextMsg(messageFile, e)

	if r.hasBlockedExtensions(body) {
		//Change Email
		return
	}

	if r.containsMalwareDomain(body) {
		//Change Email
		return
	}

	if r.hasPasswordProtectionZipFile(body) {
		//Change Email
		return
	}
}

func (r *Rules) hasPasswordProtectionZipFile(body *enmime.MIMEBody) bool {
	result := false

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= conf.MaxScanSizeKB {
			if strings.HasSuffix(fileName, ".zip") {
				content := body.Attachments[i].Content()
				result = r.isPasswordProtected(fileName, content)
			}
		}
	}

	return result
}

func (r *Rules) hasBlockedExtensions(body *enmime.MIMEBody) bool {

	result := false

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		size := binary.Size(body.Attachments[i].Content()) / 1024

		if size <= conf.MaxScanSizeKB {
			if r.hasSuffixBlocked(fileName) {
				result = true
			}
		}
	}

	return result
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

func (r *Rules) containsMalwareDomain(body *enmime.MIMEBody) bool {

	result := false
	blist := r.getBlackListDomainsFromConfig()

	for i := 0; i < len(blist); i++ {

		isMatch := strings.ContainsAny(body.HTML, blist[i])
		if isMatch {
			result = true
			break
		}
	}

	return result
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
	tmpfile := fmt.Sprintf("%v\\.tmp\\%v", conf.MePath, fileName)
	r.writeAttachmentFile(tmpfile, content)

	return isPasswordProtected(tmpfile)
}

func (r *Rules) writeAttachmentFile(fileName string, content []byte) {

	err := ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		panic(err)
	}
}

func (r *Rules) newPlainTextMsg(messageFile string, m *mail.Message) {
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

	err := ioutil.WriteFile(messageFile+".Change", content, 0644)
	if err != nil {
		panic(err)
	}
}
