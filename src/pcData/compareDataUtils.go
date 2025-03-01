package pcData

import (
	"reflect"
)

func ComparePreviousDataLogic[T any](cur T, list []T) T {
	newVal := cur
	curTest := getIdentifier(cur)
	oldVal := findOldRecord(curTest, list)
	mergePriceFields(&newVal, oldVal)

	return newVal
}

// 辅助函数：获取品牌+名称的标识符
func getIdentifier[T any](item T) string {
	v := reflect.ValueOf(item)
	brand := v.FieldByName("Brand").String()
	name := v.FieldByName("Name").String()
	return brand + name
}

// 辅助函数：查找匹配的历史记录
func findOldRecord[T any](identifier string, list []T) T {
	var zero T
	for _, item := range list {
		if getIdentifier(item) == identifier {
			return item
		}
	}
	return zero
}

// 辅助函数：合并价格字段
func mergePriceFields[T any](newVal *T, oldVal T) {
	nv := reflect.ValueOf(newVal).Elem()
	ov := reflect.ValueOf(oldVal)

	priceFields := []string{"PriceCN", "PriceUS", "PriceHK"}

	for _, field := range priceFields {
		nf := nv.FieldByName(field)
		of := ov.FieldByName(field)

		if nf.IsValid() && of.IsValid() && nf.Kind() == reflect.String {
			if nf.String() == "" {
				nf.Set(of)
			}
		}
	}
}
