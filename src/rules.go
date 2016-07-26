package main

import (
	"bufio"
	"log"
	"net/mail"
	"os"
	"strings"

	"github.com/jhillyerd/go.enmime"
)

type Rules struct {
}

func (r *Rules) ApplyRules(messageFile string) {

	e, _ := ReadEmail(messageFile)
	body, _ := enmime.ParseMIMEBody(e)

	r.clearAttachmentByExtensions(body)
	r.clearAttachmentBySize(body)

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()
		log.Printf("name: %v, content: %v", fileName, string(body.Attachments[i].Content()))
	}

}

func (r *Rules) clearAttachmentBySize(body *enmime.MIMEBody) {

	reason_msg := []byte(conf.BlockZipKBMsg)

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()

		if strings.HasSuffix(fileName, ".zip") {
			body.Attachments[i].SetContentType("text/html")
			body.Attachments[i].SetContent(reason_msg)
			body.Attachments[i].SetFileName(fileName + ".cleared")
		}
	}
}

func (r *Rules) clearAttachmentByExtensions(body *enmime.MIMEBody) {

	reason_msg := []byte(conf.BlockExtensionsMsg)

	for i := 0; i < len(body.Attachments); i++ {
		fileName := body.Attachments[i].FileName()

		if r.hasSuffixBlocked(fileName) {
			body.Attachments[i].SetContentType("text/html")
			body.Attachments[i].SetContent(reason_msg)
			body.Attachments[i].SetFileName(fileName + ".cleared")
		}
	}
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

func (r *Rules) isContainMalwareDomain(messageFile string) bool {

	result := false

	//file := fmt.Sprintf("%v\\Queues\\%v\\Inbound\\Messages\\%v", conf.MePath, connector, msgId)

	bodyTxt, err := ReadEmailBody(messageFile)

	if err != nil {
		return result
	}

	blist := r.getBlackListDomainsFromConfig()

	for i := 0; i < len(blist); i++ {

		isMatch := strings.ContainsAny(bodyTxt, blist[i])
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
