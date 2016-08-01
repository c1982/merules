package main

import (
	"fmt"
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

	messageFile := fmt.Sprintf("%v\\Queues\\%v\\Inbound\\Messages\\%v", conf.MePath, ConnectorCode, MessageID)

	var r = Rules{}
	r.config = conf
	r.ApplyRules(messageFile)

}
