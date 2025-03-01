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
	fmt.Println("LoadValidationData started!")
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

func getItemId[T any](t T) string {
	rv := reflect.Indirect(reflect.ValueOf(t))
	// 尝试常见ID字段名称
	if field := rv.FieldByName("Id"); field.IsValid() {
		return field.String()
	}
	return "unknown_id"
}

// 合并数据主逻辑
func MergeData[T any](original []T, validationData map[string][]string) []T {
	// 创建ID映射
	dataMap := make(map[string]*T)
	for i := range original {
		item := &original[i]
		id := getItemId(item)
		dataMap[id] = item
	}

	// 处理验证数据
	for id, fields := range validationData {
		item, exists := dataMap[id]
		if !exists {
			continue
		}

		v := reflect.ValueOf(item).Elem()
		for _, entry := range fields {
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) != 2 {
				continue
			}

			fieldName := strings.TrimSpace(parts[0])
			fieldValue := strings.TrimSpace(parts[1])

			field := v.FieldByName(fieldName)
			if !field.IsValid() || !field.CanSet() {
				continue
			}

			setFieldValue(field, fieldValue)
		}
	}
	return original
}

// 辅助函数：设置字段值
func setFieldValue(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintValue)
		}
	case reflect.Float32, reflect.Float64:
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatValue)
		}
	case reflect.Bool:
		if boolValue, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolValue)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			slice := make([]string, len(values))
			for i, v := range values {
				slice[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(slice))
		}
	case reflect.Ptr:
		// 简单处理指针类型（需要根据实际需求扩展）
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(&value))
		}
	}
}
