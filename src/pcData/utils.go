package pcData

import (
	"regexp"
	"strconv"
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
