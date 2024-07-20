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
