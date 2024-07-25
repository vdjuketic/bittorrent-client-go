package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func encodeBencode(data interface{}) []byte {
	switch v := data.(type) {
	case int:
		return []byte(fmt.Sprintf("i%de", data))
	case string:
		return []byte(fmt.Sprintf("%d:%s", len(fmt.Sprint(data)), data))
	case []byte:
		return v
	case []interface{}:
		return []byte(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(data)), ","), "[]"))
	case map[string]interface{}:
		res := "d"
		for key, val := range data.(map[string]interface{}) {
			res += string(encodeBencode(key))
			res += string(encodeBencode(val))
		}
		res += "e"
		return []byte(res)
	default:
		panic("unsupported data type")
	}
}

func decodeBencode(bencode string) (interface{}, error) {
	result, _, err := decodeBencodeWithDelimiter(bencode)

	return result, err
}

func decodeBencodeWithDelimiter(bencode string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencode[0])) {
		return decodeString(bencode)
	} else if rune(bencode[0]) == 'i' {
		return decodeInt(bencode)
	} else if rune(bencode[0]) == 'l' {
		return decodeList(bencode)
	} else if rune(bencode[0]) == 'd' {
		return decodeDictionary(bencode)
	} else {
		return "", 0, fmt.Errorf("invalid bencode")
	}
}

func decodeString(bencode string) (interface{}, int, error) {
	firstColonIndex := strings.Index(bencode, ":")

	lengthStr := bencode[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	endDelimeter := firstColonIndex + 1 + length
	result := bencode[firstColonIndex+1 : endDelimeter]

	return result, endDelimeter, nil
}

func decodeInt(bencode string) (interface{}, int, error) {
	eIndex := strings.Index(bencode, "e")

	result, err := strconv.Atoi(bencode[1:eIndex])
	if err != nil {
		return 0, 0, err
	}

	return result, eIndex + 1, nil
}

func decodeList(bencode string) (interface{}, int, error) {
	decodedList := make([]interface{}, 0)
	currentChar := 1

	for rune(bencode[currentChar]) != 'e' {
		decodedPart, charsRead, err := decodeBencodeWithDelimiter(bencode[currentChar:])
		if err != nil {
			return "", 0, err
		}

		decodedList = append(decodedList, decodedPart)
		currentChar += charsRead
	}

	return decodedList, currentChar + 1, nil
}

func decodeDictionary(bencode string) (interface{}, int, error) {
	result := make(map[string]interface{})
	currentChar := 1

	for rune(bencode[currentChar]) != 'e' {
		decodedKey, keyCharsRead, err := decodeBencodeWithDelimiter(bencode[currentChar:])
		if err != nil {
			return "", 0, err
		}

		currentChar += keyCharsRead

		decodedValue, valueCharsRead, err := decodeBencodeWithDelimiter(bencode[currentChar:])
		if err != nil {
			return "", 0, err
		}

		currentChar += valueCharsRead

		stringKey := fmt.Sprint(decodedKey)
		result[stringKey] = decodedValue
	}

	return result, currentChar + 1, nil
}
