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
	confTest.MePath = "O:\\Projects\\Go\\src\\merules\\mai"
	confTest.DeleteDetectedMail = false
	confTest.SenderEmail = "aspsrc@gmail.com"
	confTest.SendReportSender = false
	confTest.SendReportRecipient = false
	confTest.ScanServices = []string{"SMTP", "SF"}

	testRules = Rules{}
	testRules.config = confTest
}

var meessageFile = "O:\\Projects\\Go\\src\\merules\\mai\\1CF842E1C83B4267838199F7B5ACB0FF.MAI"
var meessageFile2 = "O:\\Projects\\Go\\src\\merules\\mai\\4AB300013D4B476E815B012A432383D6.MAI"
var meessageFile3 = "O:\\Projects\\Go\\src\\merules\\mai\\81204A08530B4C98A48C20ABA2DB80F7.MAI"
var meessageFile4 = "O:\\Projects\\Go\\src\\merules\\mai\\A8A79EED30B849D5B1767E863261BCD2.MAI"
var zipAttachment = "O:\\Projects\\Go\\src\\merules\\mai\\encrypted.zip"

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

func Test_hasPasswordProtectionZipFile(t *testing.T) {

	e, err := ReadEmail(meessageFile3)

	if err != nil {
		t.Fatal("Read Error: ", err)
	}

	body, err := enmime.ParseMIMEBody(e)

	if err != nil {
		log.Fatal("Perse MIME Error: ", err)
	}

	isIt, msg := testRules.hasPasswordProtectionZipFile(body)

	if !isIt {
		t.Error("Encrypted zip attachment not found.")
	}

	if msg != "Blocked zip: encrypted.zip" {
		t.Error("Invalid return message")
	}

}

func Test_isPasswordProtected(t *testing.T) {

	yes, err := isPasswordProtected(zipAttachment)

	if err != nil {
		log.Fatal("Zip file read error:", err)
	}

	if !yes {
		t.Error("Zip is password protected but not determine this stupid function.")
	}
}

func Test_PrintEmail(t *testing.T) {

	e, err := ReadEmail(meessageFile4)

	if err != nil {
		t.Fatal("Read Error: ", err)
	}

	emailText := ChangeEmailBodyToMessage(e, "This email cleared")

	log.Println("Mail:", emailText)
}

func Test_GetFileNameFromPath(t *testing.T) {

	fileName := getFileNameOfPath(meessageFile)

	if fileName != "1CF842E1C83B4267838199F7B5ACB0FF.MAI" {
		t.Error("Filename not determine", fileName)
	}
}

func Test_IsPermittedService(t *testing.T) {

	result := isPermittedService(confTest.ScanServices, "SMTP")

	if !result {
		t.Error("Not contains SMTP service in Permitted Services")
	}

	result = isPermittedService(confTest.ScanServices, "KAKA")

	if result {
		t.Error("Wrong behaviour IsPermittedService")
	}

}
