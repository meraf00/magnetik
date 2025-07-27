package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/meraf00/magnetik/torrent"
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := torrent.DecodeBencode([]byte(bencodedValue))
		if err != nil {
			fmt.Println(err)
			return
		}

		processed := torrent.ConvertByteToString(decoded)
		jsonOutput, _ := json.Marshal(processed)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		filePath := os.Args[2]
		content, err := os.ReadFile(filePath)

		if err != nil {
			fmt.Printf("unable to read file: %s\n%v\n", filePath, err)
			os.Exit(1)
		}

		fileInfo, err := torrent.LoadMetaInfo(content)

		if err != nil {
			os.Exit(1)
		}

		metaInfoDict := fileInfo.ToMap()
		encoded, err := torrent.EncodeBencode(map[string]any(metaInfoDict["info"].(map[string]any)))
		if err != nil {
			fmt.Println("error verifying hash")
			os.Exit(1)
		}

		hash := sha1.Sum([]byte(encoded))

		fmt.Printf("Tracker: %s\nLength: %d\nInfo Hash: %x\n", fileInfo.Announce, fileInfo.Info.Length, hash)

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
