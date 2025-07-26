package torrent

import (
	"fmt"
	"strconv"
	"unicode"
)

type Info struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      string
}

type MetaInfo struct {
	Announce string
	Info     Info
}

func decodeString(str string) (string, int, error) {
	var firstColonIndex int

	for i := 0; i < len(str); i++ {
		if str[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := str[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	if length >= len(str) {
		return "", 0, fmt.Errorf("bencode string too short")
	}

	return str[firstColonIndex+1 : firstColonIndex+1+length], firstColonIndex + length + 1, nil
}

func decodeInteger(str string) (int, int, error) {
	endIndex := len(str) - 1

	for i := 0; i < len(str); i++ {
		if str[i] == 'e' {
			endIndex = i
			break
		}
	}

	number, err := strconv.Atoi(str[1:endIndex])

	if err != nil {
		return 0, 0, err
	}

	return number, endIndex + 1, nil
}

func decodeList(str string) (interface{}, int, error) {
	currentIndex := 1

	list := []interface{}{}

	for currentIndex < len(str)-2 {
		if str[currentIndex] == 'e' {
			return list, currentIndex + 1, nil
		}

		decoded, consumedChars, err := DecodeBencode(str[currentIndex:])

		if err != nil {
			return nil, 0, err
		}

		list = append(list, decoded)
		currentIndex += consumedChars
	}

	return list, currentIndex + 1, nil
}

func decodeDict(str string) (interface{}, int, error) {
	dict := make(map[string]interface{})

	var currentIndex int = 1

	for currentIndex < len(str)-2 {
		if str[currentIndex] == 'e' {
			return dict, currentIndex + 1, nil
		}

		key, keyLength, keyErr := decodeString(str[currentIndex:])
		if keyErr != nil {
			return nil, 0, fmt.Errorf("error parsing dict key")
		}
		currentIndex += keyLength

		value, valueLength, err := DecodeBencode(str[currentIndex:])
		if err != nil {
			return nil, 0, fmt.Errorf("error parsing value for key %s %v", key, err)
		}
		currentIndex += valueLength

		dict[key] = value
	}

	return dict, currentIndex + 1, nil
}

func DecodeBencode(bencodedString string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInteger(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else if bencodedString[0] == 'd' {
		return decodeDict(bencodedString)
	} else {
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
}

func LoadMetaInfo(bencodedString string) (*MetaInfo, error) {
	decoded, _, decodingError := DecodeBencode(bencodedString)

	if decodingError != nil {
		return nil, decodingError
	}

	metaInfo, ok := decoded.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid meta info, corrupt torrent file")
	}

	info, ok := metaInfo["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid info, corrupt torrent file")
	}

	return &MetaInfo{
		Announce: metaInfo["announce"].(string),
		Info: Info{
			Length:      info["length"].(int),
			Name:        info["name"].(string),
			PieceLength: info["piece length"].(int),
			Pieces:      info["pieces"].(string),
		},
	}, nil
}
