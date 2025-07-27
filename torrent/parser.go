package torrent

import (
	"fmt"
	"sort"
	"strconv"
	"unicode"
)

type Info struct {
	Length      int    `json:"length"`
	Name        string `json:"name"`
	PieceLength int    `json:"piece length"`
	Pieces      []byte `json:"pieces"`
}

type MetaInfo struct {
	Announce string `json:"announce"`
	Info     Info   `json:"info"`
}

func (m *MetaInfo) ToMap() map[string]any {
	return map[string]any{
		"announce": m.Announce,
		"info": map[string]any{
			"length":       m.Info.Length,
			"name":         m.Info.Name,
			"piece length": m.Info.PieceLength,
			"pieces":       m.Info.Pieces,
		},
	}
}

func decodeString(str []byte) ([]byte, int, error) {
	var firstColonIndex int

	for i := range str {
		if str[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := str[:firstColonIndex]

	length, err := strconv.Atoi(string(lengthStr))
	if err != nil {
		return nil, 0, err
	}

	if length >= len(str) {
		return nil, 0, fmt.Errorf("bencode string too short")
	}

	return str[firstColonIndex+1 : firstColonIndex+1+length], firstColonIndex + length + 1, nil
}

func decodeInteger(str []byte) (int, int, error) {
	endIndex := len(str) - 1

	for i := 0; i < len(str); i++ {
		if str[i] == 'e' {
			endIndex = i
			break
		}
	}

	number, err := strconv.Atoi(string(str[1:endIndex]))

	if err != nil {
		return 0, 0, err
	}

	return number, endIndex + 1, nil
}

func decodeList(str []byte) (interface{}, int, error) {
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

func decodeDict(str []byte) (interface{}, int, error) {
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

		dict[string(key)] = value
	}

	return dict, currentIndex + 1, nil
}

func DecodeBencode(bencodedString []byte) (any, int, error) {
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

func EncodeBencode(obj any) ([]byte, error) {
	switch obj := obj.(type) {
	case int:
		return fmt.Appendf(nil, "i%de", obj), nil
	case string:
		return fmt.Appendf(nil, "%d:%s", len(obj), obj), nil
	case []byte:
		return fmt.Appendf(nil, "%d:%s", len(obj), obj), nil
	case []any:
		encoded := make([]byte, 0)
		encoded = append(encoded, 'l')

		for _, element := range obj {
			encodedElement, err := EncodeBencode(element)
			if err != nil {
				return nil, err
			}

			encoded = append(encoded, encodedElement...)
		}

		encoded = append(encoded, 'e')
		return encoded, nil
	case map[string]any:
		encoded := make([]byte, 0)
		encoded = append(encoded, 'd')

		keys := make([]string, 0)
		for key := range obj {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, k := range keys {
			encodedKey, err := EncodeBencode(k)
			if err != nil {
				return nil, err
			}

			encoded = append(encoded, encodedKey...)

			encodedValue, err := EncodeBencode(obj[k])
			if err != nil {
				return nil, err
			}

			encoded = append(encoded, encodedValue...)
		}

		encoded = append(encoded, 'e')
		return encoded, nil
	}

	return nil, fmt.Errorf("error encoding object")
}

func LoadMetaInfo(bencodedString []byte) (*MetaInfo, error) {
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
		Announce: string(metaInfo["announce"].([]byte)),
		Info: Info{
			Length:      info["length"].(int),
			Name:        string(info["name"].([]byte)),
			PieceLength: info["piece length"].(int),
			Pieces:      info["pieces"].([]byte),
		},
	}, nil
}

func ConvertByteToString(v any) any {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case map[string]any:
		m := make(map[string]any)
		for k, v2 := range val {
			m[k] = ConvertByteToString(v2)
		}
		return m
	case []any:
		for i := range val {
			val[i] = ConvertByteToString(val[i])
		}
		return val
	default:
		return v
	}
}
