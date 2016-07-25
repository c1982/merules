package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jhillyerd/go.enmime"
)

type Rules struct {
}

func (r *Rules) IsExtensionMaxLenghtBelow(msgId string, connector string) bool {
	result := false

	return result
}

func (r *Rules) IsContainsMalwareDomain(msgId string, connector string) bool {

	result := false

	file := fmt.Sprintf("%v\\Queues\\%v\\Inbound\\Messages\\%v", conf.MePath, connector, msgId)
	//metaFile := fmt.Sprintf("%v\\Queues\\%v\\Inbound\\%v", conf.MePath, connector, msgId)

	bodyTxt := r.getEmailBody(file)
	blist := r.getBlackList()

	for i := 0; i < len(blist); i++ {

		isMatch := strings.ContainsAny(bodyTxt, blist[i])
		if isMatch {
			result = true
			break
		}
	}

	return result
}

func (r *Rules) getEmailBody(file string) string {
	msg, err := ReadEmail(file)
	mimes, err := enmime.ParseMIMEBody(msg)

	body, err := ioutil.ReadAll(msg.Body)

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s", body)
}

func (r *Rules) getBlackList() []string {

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
