package main

import (
	"encoding/json"
	"fmt"
	"go-colly-lib/src/pcData"
	"os"
	"strings"
)

// 验证函数
func validateCases(cases []pcData.CaseType) map[string][]string {
	errors := make(map[string][]string)

	for _, c := range cases {
		var errs []string

		if c.CaseSize == "" {
			errs = append(errs, "CaseSize:")
		}
		if len(c.Compatibility) == 0 {
			errs = append(errs, "Compatibility:")
		}
		if c.MaxVGAlength == 0 {
			errs = append(errs, "MaxVGAlength:")
		}
		if c.RadiatorSupport == 0 {
			errs = append(errs, "RadiatorSupport:")
		}
		if c.MaxCpuCoolorHeight == 0 {
			errs = append(errs, "MaxCpuCoolorHeight:")
		}

		if len(errs) > 0 {
			errors[c.Id] = errs
		}
	}
	return errors
}

func main() {
	// 读取输入文件
	cases := loadData("../tmp/caseData.json")

	// 加载已有验证结果
	existingErrors := loadExistingErrors("../tmp/validation/caseValidation.json")

	// 执行验证
	newErrors := validateCases(cases)

	// 智能合并结果
	finalErrors := mergeErrors(existingErrors, newErrors)

	// 生成错误报告
	saveErrors("caseValidation.json", finalErrors)

	fmt.Println("Validation completed. Errors saved to caseValidation.json")
}

func mergeErrors(old, new map[string][]string) map[string][]string {
	result := make(map[string][]string)

	// 复制旧数据
	for id, errs := range old {
		result[id] = append([]string{}, errs...)
	}

	// 合并新数据
	for id, newErrs := range new {
		existing := result[id]
		existingFields := make(map[string]bool)

		// 记录现有字段
		for _, err := range existing {
			existingFields[extractField(err)] = true
		}

		// 添加新错误
		for _, err := range newErrs {
			field := extractField(err)
			if !existingFields[field] {
				existing = append(existing, err)
				existingFields[field] = true
			}
		}

		result[id] = existing
	}
	return result
}

// 辅助函数：提取字段名称
func extractField(errorMsg string) string {
	return strings.Split(errorMsg, ":")[0]
}

// 数据加载函数
func loadData(path string) []pcData.CaseType {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var cases []pcData.CaseType
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cases); err != nil {
		panic(err)
	}
	return cases
}

// 错误数据加载
func loadExistingErrors(path string) map[string][]string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string][]string)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var errors map[string][]string
	if err := json.Unmarshal(file, &errors); err != nil {
		panic(err)
	}
	return errors
}

// 保存结果
func saveErrors(path string, errors map[string][]string) {
	output, err := json.MarshalIndent(errors, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("../tmp/validation/"+path, output, 0644); err != nil {
		panic(err)
	}
}
