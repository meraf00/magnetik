package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/meraf00/magnetik/torrent"
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := torrent.DecodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filePath := os.Args[2]
		content, err := os.ReadFile(filePath)

		if err != nil {
			fmt.Printf("unable to read file: %s\n%v\n", filePath, err)
			os.Exit(1)
		}

		fileInfo, err := torrent.LoadMetaInfo(string(content))

		if err != nil {
			os.Exit(1)
		}

		fmt.Printf("Tracker: %s\nLength: %d", fileInfo.Announce, fileInfo.Info.Length)

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
