package main

import (
	"log"
	"testing"

	"github.com/jhillyerd/go.enmime"
)

var confTest meConfig
var testRules Rules

func init() {

	confTest = meConfig{}
	confTest.ScanMalwareDomain = true
	confTest.ScanMalwareDomainMsg = "Is Mailware Domain: %1"
	confTest.MaxScanSizeKB = 512
	confTest.BlockPassZip = true
	confTest.BlockPassZipMsg = "Blocked zip: %1"
	confTest.BlockExtensions = []string{"exe", "bat", "xls"}
	confTest.BlockExtensionsMsg = "Blocked extensions: %1"
	confTest.ScanMalwareDomain = true
	confTest.ScanMalwareDomainMsg = "Is malware: %1"

	testRules = Rules{}
	testRules.config = confTest

}

var meessageFile = "O:\\Projects\\Go\\src\\merules\\mai\\1CF842E1C83B4267838199F7B5ACB0FF.MAI"
var meessageFile2 = "O:\\Projects\\Go\\src\\merules\\mai\\4AB300013D4B476E815B012A432383D6.MAI"

func Test_hasSuffixBlocked(t *testing.T) {

	if !testRules.hasSuffixBlocked("file.exe") {
		t.Error("exe file not blocked")
	}

	if testRules.hasSuffixBlocked("file.docx") {
		t.Error("docx is allowed extension.")
	}
}

func Test_getBlackListDomainsFromConfig(t *testing.T) {

	var blacklist = testRules.getBlackListDomainsFromConfig()

	if len(blacklist) == 0 {
		t.Error("Black list is empty!")
	}
}

func Test_isContainsBody(t *testing.T) {

	var blacklist = testRules.getBlackListDomainsFromConfig()
	var bodyText = `Lorem Ipsum, dizgi ve baskı endüstrisinde kullanılan mıgır metinlerdir. 
					noktafatura.net Lorem Ipsum, adı bilinmeyen bir matbaacının bir hurufat numune 
					kitabı oluşturmak üzere bir yazı galerisini alarak karıştırdığı 1500'lerden beri 
					endüstri standardı sahte metinler olarak kullanılmıştır. Beşyüz yıl boyunca varlığını 
					sürdürmekle kalmamış, aynı zamanda pek değişmeden elektronik dizgiye de sıçramıştır.
					1960'larda Lorem Ipsum pasajları da içeren`

	result, msg := testRules.isContainsBody(bodyText, blacklist)

	if !result {
		t.Error("Malware domain not found in bodyText")
	}

	if msg != "Is malware: noktafatura.net" {
		t.Log(msg)
		t.Error("Invalid return message")
	}
}

func Test_hasBlockedExtensions(t *testing.T) {

	e, err := ReadEmail(meessageFile)

	if err != nil {
		t.Fatal("Read Error: ", err)
	}

	body, err := enmime.ParseMIMEBody(e)

	if err != nil {
		log.Fatal("Perse MIME Error: ", err)
	}

	isIt, msg := testRules.hasBlockedExtensions(body)

	if isIt {
		t.Error("blocked extensions found.")
	}

	if msg != "Blocked extensions: %1" {
		t.Error("Invalid return message")
	}
}

func Test_hasBlockedExtensions_XLS(t *testing.T) {

	e, err := ReadEmail(meessageFile2)

	if err != nil {
		t.Fatal("Read Error: ", err)
	}

	body, err := enmime.ParseMIMEBody(e)

	if err != nil {
		log.Fatal("Perse MIME Error: ", err)
	}

	isIt, msg := testRules.hasBlockedExtensions(body)

	if !isIt {
		t.Error("blocked extensions not found.")
	}

	if msg != "Blocked extensions: Haftalik_Yapilacaklar_Listesi.xls" {
		t.Error("Invalid return message")
	}
}
