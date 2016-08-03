package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var conf meConfig

type meConfig struct {
	MaxScanSizeKB        int      `toml:"MaxScanSizeKB"`
	BlockPassZip         bool     `toml:"BlockZipEncrypted"`
	BlockPassZipMsg      string   `toml:"BlockPassZip_Msg"`
	BlockExtensions      []string `toml:"BlockExtensions"`
	BlockExtensionsMsg   string   `toml:"BlockExtensions_Msg"`
	ScanMalwareDomain    bool     `toml:"ScanMalwareDomain"`
	ScanMalwareDomainMsg string   `toml:"ScanMalwareDomain_Msg"`
	EmailFooter          string   `toml:"EmailFooter"`
	MePath               string   `toml:"MailEnablePath"`
	InboundScan          bool     `toml:"InboundScan"`
	OutboundScan         bool     `toml:"OutboundScan"`
	DeleteDetectedMail   bool     `toml:"DeleteDetectedMail"`
	ScanServices         []string `toml:"ScanServices"`
	SendReportRecipient  bool     `toml:"SendReportRecipient"`
	SendReportSender     bool     `toml:"SendReportSender"`
	SenderEmail          string   `toml:"SenderEmail"`
}

func init() {

	configDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	configFile := fmt.Sprintf("%v\\merules.config", configDir)

	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		if err != nil {
			fmt.Println("Failed to parse toml data: ", err)
			os.Exit(1)
		}
	}

}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("MailEnable MTA Pickup Event by MaestroPanel")
		fmt.Println("https://github.com/c1982/merules - aspsrc@gmail.com")
		return
	}

	MessageID := os.Args[1]
	ConnectorCode := os.Args[2]

	if !isPermittedService(ConnectorCode) {
		log.Println("Not permitted for scan:", ConnectorCode)
		return
	}

	inboundMessage := fmt.Sprintf("%v\\Queues\\%v\\Inbound\\Messages\\%v", conf.MePath, ConnectorCode, MessageID)
	outboundMessage := fmt.Sprintf("%v\\Queues\\%v\\Outgoing\\Messages\\%v", conf.MePath, ConnectorCode, MessageID)

	var r = Rules{}
	r.config = conf

	if conf.InboundScan {
		r.ApplyRules(inboundMessage)
	}

	if conf.OutboundScan {
		r.ApplyRules(outboundMessage)
	}

}
