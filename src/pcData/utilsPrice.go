package pcData

import (
	"fmt"
	"sort"
	"strconv"
)

const (
	Platform_JD     string = "JD"
	Platform_Taobao string = "JB"
	Platform_Newegg string = "Newegg"
	Platform_Amazon string = "Amazon"
)

// Handle the first CN price
func handleFirstCNPrice(record LinkRecord) PriceType {
	if record.PriceCN != "" {
		return PriceType{
			Region:    "CN",
			Platform:  Platform_Taobao,
			PriceLink: record.LinkCN,
			Price:     record.PriceCN,
		}
	}
	return PriceType{
		Region:    "CN",
		Platform:  Platform_JD,
		PriceLink: record.LinkCN,
		Price:     "",
	}
}

func handleSpecPricesLogic(prices []PriceType, record LinkRecord) []PriceType {
	if record.LinkCN != "" {
		fmt.Println("LinkCN:", record.LinkCN)
		prices = append(prices, handleFirstCNPrice(record))
	}
	if record.LinkUS != "" {
		prices = upsertPrice(prices, PriceType{
			Region:    "US",
			Platform:  "Newegg",
			PriceLink: record.LinkUS,
		})
	}
	return prices
}

// 輔助函數：根據Region+Platform獲取價格
func getPriceByPlatform(prices []PriceType, region, platform string) *PriceType {
	for _, p := range prices {
		if p.Region == region && p.Platform == platform {
			return &p
		}
	}
	return nil
}

// 輔助函數：更新或插入價格數據（使用Region+Platform作為唯一鍵）
func upsertPrice(prices []PriceType, newPrice PriceType) []PriceType {
	for i, p := range prices {
		if p.Region == newPrice.Region && p.Platform == newPrice.Platform {
			// 保留已有價格數據除非新數據有更新
			if newPrice.Price != "" {
				if newPrice.Platform == Platform_Newegg {
					if p.Price == "" {
						prices[i].Price = newPrice.Price
						prices[i].PriceLink = newPrice.PriceLink
					}
				} else {
					prices[i].Price = newPrice.Price
					prices[i].PriceLink = newPrice.PriceLink
				}
			}
			return prices
		}
	}
	return append(prices, newPrice)
}

type priceSortItem struct {
	price     float64
	isValid   bool
	priceLink string
	original  PriceType
}

// 輔助函數：價格數據去重
func deduplicatePrices(prices []PriceType) []PriceType {
	priceMap := make(map[string][]PriceType)

	// 1. 按Region+Platform分組
	for _, p := range prices {
		key := fmt.Sprintf("%s|%s", p.Region, p.Platform)
		priceMap[key] = append(priceMap[key], p)
	}

	result := make([]PriceType, 0)
	for _, group := range priceMap {
		if len(group) == 1 {
			result = append(result, group[0])
			continue
		}

		// 2. 過濾有效價格
		var validPrices []PriceType
		var invalidPrices []PriceType
		for _, p := range group {
			if price, err := strconv.ParseFloat(p.Price, 64); err == nil && price > 0 {
				validPrices = append(validPrices, p)
			} else {
				invalidPrices = append(invalidPrices, p)
			}
		}

		// 3. 按規則選擇
		switch {
		case len(validPrices) > 0:
			// 按價格升序排序
			sort.Slice(validPrices, func(i, j int) bool {
				pi, _ := strconv.ParseFloat(validPrices[i].Price, 64)
				pj, _ := strconv.ParseFloat(validPrices[j].Price, 64)
				return pi < pj
			})
			result = append(result, validPrices[0])
		case len(invalidPrices) > 0:
			// 保留最舊的數據（按PriceLink排序）
			sort.Slice(invalidPrices, func(i, j int) bool {
				return invalidPrices[i].PriceLink < invalidPrices[j].PriceLink
			})
			result = append(result, invalidPrices[0])
		}
	}

	sortItems := make([]priceSortItem, len(result))
	for i, p := range result {
		price, err := strconv.ParseFloat(p.Price, 64)
		sortItems[i] = priceSortItem{
			price:     price,
			isValid:   err == nil && price > 0,
			priceLink: p.PriceLink,
			original:  p,
		}
	}

	// 排序逻辑
	sort.Slice(sortItems, func(i, j int) bool {
		switch {
		case sortItems[i].isValid && sortItems[j].isValid:
			return sortItems[i].price < sortItems[j].price
		case !sortItems[i].isValid && !sortItems[j].isValid:
			return sortItems[i].priceLink < sortItems[j].priceLink
		default:
			return sortItems[i].isValid
		}
	})

	// 重建结果
	final := make([]PriceType, len(sortItems))
	for i, item := range sortItems {
		final[i] = item.original
	}

	return final
}
