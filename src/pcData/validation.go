package pcData

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func LoadValidationData(path string) map[string][]string {
	file, err := os.ReadFile("tmp/validation/" + path + "Validation.json")
	if err != nil {
		fmt.Println("加载验证数据失败:", err)
	}

	var validationData map[string][]string
	if err := json.Unmarshal(file, &validationData); err != nil {
		fmt.Println("加载验证数据失败:", err)
	}

	return validationData
}

// 合并数据主逻辑
func MergeCases(original []CaseType, validationData map[string][]string) []CaseType {
	// 创建ID到Case的映射以便快速查找
	caseMap := make(map[string]*CaseType)
	for i := range original {
		caseMap[original[i].Id] = &original[i]
	}

	// 遍历验证数据
	for caseID, fields := range validationData {
		currentCase, exists := caseMap[caseID]
		if !exists {
			continue // 忽略不存在的Case
		}

		// 处理每个字段的更新
		for _, fieldEntry := range fields {
			parts := strings.SplitN(fieldEntry, ":", 2)
			if len(parts) != 2 {
				continue // 无效格式
			}

			fieldName := strings.TrimSpace(parts[0])
			fieldValue := strings.TrimSpace(parts[1])

			// 使用反射进行安全赋值
			rv := reflect.ValueOf(currentCase).Elem()
			field := rv.FieldByName(fieldName)
			if !field.IsValid() {
				continue // 忽略不存在的字段
			}

			// 根据字段类型转换值
			switch field.Kind() {
			case reflect.String:
				field.SetString(fieldValue)
			case reflect.Int:
				if intValue, err := strconv.Atoi(fieldValue); err == nil {
					field.SetInt(int64(intValue))
				}
			case reflect.Slice: // 处理Compatibility字段
				if field.Type().Elem().Kind() == reflect.String {
					values := strings.Split(fieldValue, ",")
					for i := range values {
						values[i] = strings.TrimSpace(values[i])
					}
					field.Set(reflect.ValueOf(values))
				}
			}
		}
	}
	return original
}
