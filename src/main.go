package main

func main() {

	if len(os.Args) < 2 {
		fmt.Println("MaestroPanel MTA Filter")
		return
	}

	MessageID := os.Args[1]
	ConnectorCode := os.Args[2]

	log.Println("Message ID:", MessageID, "Connector:", ConnectorCode)

}
