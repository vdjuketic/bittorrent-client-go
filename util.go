package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func decodeBencode(bencodedValue string) string {
	decodedValue, _ := decodeBencodeWithEndDelimeter(bencodedValue)
	return decodedValue
}

func decodeBencodeWithEndDelimeter(bencodedValue string) (string, int) {
	firstChar := bencodedValue[0]

	if firstChar >= '0' && firstChar <= '9' {
		return decodeString(bencodedValue)
	} else if firstChar == 'i' {
		return decodeInt(bencodedValue)
	} else if firstChar == 'l' {
		return decodeList(bencodedValue)
	} else if firstChar == 'd' {
		return decodeDictionary(bencodedValue)
	} else {
		fmt.Printf("Invalid bencoded value.")
		os.Exit(1)
	}
}

func decodeString(bencodedValue string) (string, int) {
	colonIndex := strings.Index(bencodedValue, ":")

	if colonIndex == -1 {
		fmt.Printf("Invalid bencoded value.")
		os.Exit(1)
	}

	length, err := strconv.Atoi(bencodedValue[:colonIndex])
	if err != nil {
		fmt.Printf("Invalid bencoded value.")
		panic(err)
	}

	endDelimeter := colonIndex + 1 + length

	decodedString := bencodedValue[colonIndex+1 : endDelimeter]

	return decodedString, endDelimeter
}

func decodeInt(bencodedValue string) (int, int) {
	endDelimeter := strings.Index(bencodedValue, "e")
	bencodedValue = bencodedValue[1:endDelimeter]

	decodedInt, err := strconv.Atoi(strings.Replace(bencodedValue, "~", "-", -1))
	if err != nil {
		fmt.Printf("Invalid bencoded value.")
		panic(err)
	}

	return decodedInt, endDelimeter + 1
}

func decodeList(bencodedValue string) {

}

func decodeDictionary(bencodedValue string) {

}
