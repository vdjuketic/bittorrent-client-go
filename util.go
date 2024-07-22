package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func decodeBencode(bencodedValue string) string {
	decodedValue, _ := decodeBencodeWithEndDelimeter(bencodedValue)
	return decodedValue
}

func decodeBencodeWithEndDelimeter(bencodedValue string) (string, int) {
	firstChar := bencodedValue[0]

	if unicode.IsDigit(rune(firstChar)) && firstChar >= '0' && firstChar <= '9' {
		return decodeString(bencodedValue)
	} else if firstChar == 'i' {
		return decodeInt(bencodedValue)
	} else if firstChar == 'l' {
		return decodeList(bencodedValue)
	} else if firstChar == 'd' {
		return decodeDictionary(bencodedValue)
	}
	fmt.Println("Invalid bencoded value.")
	os.Exit(1)
	return "", 0
}

func decodeString(bencodedValue string) (string, int) {
	colonIndex := strings.Index(bencodedValue, ":")

	if colonIndex == -1 {
		fmt.Println("Invalid bencoded string.")
		os.Exit(1)
	}

	length, err := strconv.Atoi(bencodedValue[:colonIndex])
	if err != nil {
		fmt.Println("Invalid bencoded string.")
		panic(err)
	}

	endDelimeter := colonIndex + 1 + length

	decodedString := bencodedValue[colonIndex+1 : endDelimeter]

	return decodedString, endDelimeter
}

func decodeInt(bencodedInt string) (string, int) {
	endDelimeter := strings.Index(bencodedInt, "e")
	bencodedInt = bencodedInt[1:endDelimeter]

	decodedInt := strings.Replace(bencodedInt, "~", "-", -1)

	return decodedInt, endDelimeter + 1
}

func decodeList(bencodedList string) (string, int) {
	decodedList := make([]string, 0)
	currentChar := 1

	for {
		if bencodedList[currentChar] != 'e' {
			break
		}

		decodedPart, charsRead := decodeBencodeWithEndDelimeter(bencodedList[currentChar:])
		decodedList = append(decodedList, decodedPart)
		currentChar += charsRead
	}

	return strings.Join(decodedList, ","), currentChar + 1
}

func decodeDictionary(bencodedDictionary string) (string, int) {
	decodedDictionary := make(map[string]string)
	currentChar := 1

	for {
		if bencodedDictionary[currentChar] == 'e' {
			break
		}

		decodedKey, keyCharsRead := decodeBencodeWithEndDelimeter(bencodedDictionary[currentChar:])
		currentChar += keyCharsRead

		decodedValue, valueCharsRead := decodeBencodeWithEndDelimeter(bencodedDictionary[currentChar:])
		currentChar += valueCharsRead

		decodedDictionary[decodedKey] = decodedValue
	}

	stringRepresentation, err := json.Marshal(decodedDictionary)
	if err != nil {
		fmt.Println("Invalid bencoded dictionary.")
		panic(err)
	}

	return string(stringRepresentation), currentChar + 1
}
