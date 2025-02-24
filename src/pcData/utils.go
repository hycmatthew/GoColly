package pcData

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/gocolly/colly/v2"
)

func checkPriceValid(str string) bool {
	return str != ""
}

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

func strContains(s string, sub string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(sub))
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

func OutOfStockLogic(usPrice string, str string) string {
	if strings.Contains(str, "Out of Stock") {
		return "Out of Stock"
	}
	return usPrice
}

func GetPriceLinkFromPangoly(element *colly.HTMLElement) (string, string) {
	resPrice := "Out of Stock"
	resLink := ""
	loopBreak := false

	element.ForEach("table.table-prices tr", func(i int, item *colly.HTMLElement) {
		if !loopBreak {
			tempPrice := extractFloatStringFromString(item.ChildText(".detail-purchase strong"))
			// tempAvailability := item.ChildText(".hidden-xs span")
			tempLink := item.ChildAttr(".detail-purchase", "href")

			if strings.Contains(tempLink, "amazon") {
				amazonLink := strings.Split(tempLink, "?tag=")[0]
				resLink = amazonLink
				resPrice = tempPrice
			}
			if strings.Contains(tempLink, "newegg") {
				neweggLink := strings.Split(tempLink, "url=")[1]
				UnescapeLink, _ := url.QueryUnescape(neweggLink)
				neweggLink = strings.Split(UnescapeLink, "\u0026")[0]
				resLink = neweggLink
				resPrice = tempPrice
				loopBreak = true
			}
		}
	})

	return resPrice, resLink
}

func SetProductId(brand string, name string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9 -]+")
	tempStr := re.ReplaceAllString(brand+"-"+name, "")
	result := strings.ToLower(strings.ReplaceAll(tempStr, " ", "-"))
	return MergeDashes(result)
}

func RemoveBrandsFromName(brand, name string) string {
	pattern := "(?i)" + regexp.QuoteMeta(brand)
	re := regexp.MustCompile(pattern)

	count := 0
	return re.ReplaceAllStringFunc(name, func(matched string) string {
		if count < 1 {
			count++
			return ""
		}
		return matched
	})
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(input []string) []string {
	// Create a map to track unique strings
	uniqueMap := make(map[string]struct{})
	var result []string

	for _, str := range input {
		if _, exists := uniqueMap[str]; !exists {
			uniqueMap[str] = struct{}{}  // Add unique string to the map
			result = append(result, str) // Add to result slice
		}
	}

	return result
}

func MergeDashes(s string) string {
	var builder strings.Builder
	prevDash := false

	for _, r := range s {
		if r == '-' {
			if !prevDash {
				builder.WriteRune('-')
				prevDash = true
			}
		} else {
			builder.WriteRune(r)
			prevDash = false
		}
	}
	return builder.String()
}

// Merge Data
// 通用合并函数
func mergeValue(id string, v1, v2 interface{}) interface{} {
	// 统一转换为反射值
	rv1 := reflect.ValueOf(v1)
	rv2 := reflect.ValueOf(v2)

	// 类型必须一致
	if rv1.Type() != rv2.Type() {
		panic("mergeValue: type mismatch")
	}

	// 判断是否双方都有数据
	hasV1 := !isEmpty(rv1)
	hasV2 := !isEmpty(rv2)

	// 记录数据对比
	if hasV1 && hasV2 {
		if v1 != v2 {
			fmt.Printf("[CONFLICT] %s:\n  V1 = %s\n  V2 = %s\n", id, rv1, rv2)
		}
	}

	// 判断v1是否为空/零值
	if isEmpty(rv1) {
		return v2
	}
	return v1
}

// 判断值是否为空/零值
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Interface:
		return v.IsNil() || isEmpty(v.Elem())
	default:
		return v.IsZero()
	}
}

// 自动获取ID字段（支持常见命名方式）
func getID(v interface{}) string {
	rv := reflect.Indirect(reflect.ValueOf(v))

	// 尝试常见ID字段名称
	for _, fieldName := range []string{"Code", "Name"} {
		if field := rv.FieldByName(fieldName); field.IsValid() {
			return fmt.Sprintf("%v", field)
		}
	}
	return "unknown_id"
}

func MergeStruct(s1, s2 interface{}) interface{} {
	rv1 := reflect.ValueOf(s1)
	rv2 := reflect.ValueOf(s2)

	// 创建新结构体
	result := reflect.New(rv1.Type()).Elem()
	tempId := getID(s1)

	// 遍历字段
	for i := 0; i < rv1.NumField(); i++ {
		field1 := rv1.Field(i)
		field2 := rv2.Field(i)

		// 递归合并每个字段
		merged := mergeValue(tempId, field1.Interface(), field2.Interface())
		result.Field(i).Set(reflect.ValueOf(merged))
	}

	return result.Interface()
}
