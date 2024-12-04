package pcData

import (
	"regexp"
	"strconv"
	"strings"
)

func extractNumberFromString(str string) int {
	digitCheck := regexp.MustCompile("[0-9]")
	re := regexp.MustCompile("[0-9]+")
	if digitCheck.MatchString(str) {
		i, err := strconv.Atoi(re.FindAllString(strings.Replace(str, ",", "", -1), -1)[0])
		if err != nil {
			panic(err)
		}
		return i
	}
	return 0
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
	return strings.TrimSpace(result)
}

func isSubset(shortArr, longArr []string) bool {
	set := make(map[string]bool)

	for _, str := range longArr {
		set[str] = true
	}

	for _, str := range shortArr {
		if !set[str] {
			return false
		}
	}
	return true
}

func getWordBeforeSpecificString(input string, specificString string) string {
	index := strings.Index(input, specificString)
	if index == -1 {
		return ""
	}
	resList := strings.Split(strings.TrimSpace(input[:index]), " ")
	return resList[len(resList)-1]
}

func getStringBeforeSpecificString(input string, specificString string) string {
	index := strings.Index(input, specificString)
	if index == -1 {
		return ""
	}
	return strings.TrimSpace(input[:index])
}
