package pcData

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/axgle/mahonia"
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

func extractEngCharFromString(str string) string {
	result := regexp.MustCompile(`[^a-zA-Z0-9-]+`).ReplaceAllString(str, "")
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

func convertGBKString(str string) string {
	encoder := mahonia.NewEncoder("GBK")
	gbkBytes := encoder.ConvertString(str)
	return string(gbkBytes)
}

func removeElement(nums []string, val string) []string {
	for i := 0; i < len(nums); {
		if nums[i] == val {
			if len(nums) > i {
				nums = append(nums[:i], nums[i+1:]...)
				continue
			}
		}
		i++
	}
	return nums
}
func getTheLargestValueInArray(num []string) int {
	res := 0
	for _, element := range num {
		testVal := extractNumberFromString(element)
		if testVal > res {
			res = testVal
		}
	}
	return res
}

func SplitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

func SocketContainLogic(strList []string) []string {
	socketList := []string{
		"1150", "1151", "1155", "1156", "1200", "1700", "1851", "2011", "2011-3", "2066", "AM2", "AM3", "AM4", "AM5", "AM2+", "AM3+", "FM1", "FM2", "FM2+",
	}
	var resultList []string

	for _, element := range socketList {
		for _, testStr := range strList {
			if strings.Contains(testStr, element) {
				resultList = append(resultList, element)
			}
		}
	}

	for _, testStr := range strList {
		if strings.Contains(testStr, "115X") {
			resultList = append(resultList, []string{"1150", "1151", "1155", "1156"}...)
		}
	}
	return resultList
}
