package pcData

import (
	"regexp"
	"strconv"
	"strings"
)

func extractNumberFromString(str string) int {
	re := regexp.MustCompile("[0-9]+")
	i, err := strconv.Atoi(re.FindAllString(str, -1)[0])
	if err != nil {
		panic(err)
	}

	return i
}

func extractFloatStringFromString(str string) string {
	re := regexp.MustCompile(`[\d.]+`)
	matches := re.FindAllString(str, -1)
	result := ""

	if len(matches) > 0 {
		result = matches[0]
	}

	return result
}

func extractGPUStringFromString(str string) string {
	splitedStr := strings.Split(str, "NVIDIA GeForce")
	result := ""
	if len(splitedStr) > 1 {
		result = splitedStr[1]
	}
	return result
}
