package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

var conf meConfig

type meConfig struct {
	BlockZipKB           int      `toml:"BlockZipKB"`
	BlockZipKBMsg        string   `toml:"BlockZipKB_Msg"`
	BlockZipEncrypted    bool     `toml:"BlockZipEncrypted"`
	BlockZipEncryptedMsg string   `toml:"BlockZipEncrypted_Msg"`
	BlockExtensions      []string `toml:"BlockExtensions"`
	BlockExtensionsMsg   string   `toml:"BlockExtensions_Msg"`
	MePath               string   `toml:"MailEnablePath"`
}

func init() {

	if _, err := toml.DecodeFile("./merules.config", &conf); err != nil {
		if err != nil {
			fmt.Println("Failed to parse toml data: ", err)
			os.Exit(1)
		}
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("MaestroPanel MTA Filter")
		return
	}

	MessageID := os.Args[1]
	ConnectorCode := os.Args[2]

	log.Println("Message ID:", MessageID, "Connector:", ConnectorCode)

}
