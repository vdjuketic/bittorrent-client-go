package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func encodePair(key, value interface{}) ([]byte, []byte, error) {
	encodedKey, err := encodeBencode(key)
	if err != nil {
		return nil, nil, err
	}

	encodedValue, err := encodeBencode(value)
	if err != nil {
		return nil, nil, err
	}

	return encodedKey, encodedValue, nil
}

func encodeBencode(data interface{}) ([]byte, error) {
	var encoded string

	switch t := data.(type) {
	case string:
		encoded = fmt.Sprintf("%d:%s", len(t), t)
	case int:
		encoded = fmt.Sprintf("i%de", t)
	case []interface{}:
		encodedElements := make([]string, len(t))

		for i, val := range t {
			encodedElement, err := encodeBencode(val)
			if err != nil {
				return nil, err
			}
			encodedElements[i] = string(encodedElement)
		}

		sort.Strings(encodedElements)
		encodedList := strings.Join(encodedElements, "")
		encoded = fmt.Sprintf("l%se", encodedList)

	case map[string]interface{}:
		sortedKeys := make([]string, 0, len(t))
		for key := range t {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sort.StringSlice(sortedKeys))

		encodedDictionary := make([]string, 0, len(t)*2)
		for _, key := range sortedKeys {
			encodedKey, encodedVal, err := encodePair(key, t[key])
			if err != nil {
				return nil, err
			}
			encodedDictionary = append(encodedDictionary, string(encodedKey), string(encodedVal))
		}
		encoded = fmt.Sprintf("d%se", strings.Join(encodedDictionary, ""))
	default:
		return nil, fmt.Errorf("data type not supported %v", t)
	}
	return []byte(encoded), nil
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
