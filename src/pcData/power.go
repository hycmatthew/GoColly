package pcData

import (
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type PowerSpec struct {
	Code        string
	Brand       string
	Name        string
	ReleaseDate string
	Wattage     int
	Size        string
	Standard    string
	Modular     string
	Efficiency  string
	Length      int
	Prices      []PriceType
	Img         string
}

type PowerType struct {
	Id          string
	Brand       string
	Name        string
	NameCN      string
	ReleaseDate string
	Wattage     int
	Size        string
	Standard    string
	Modular     string
	Efficiency  string
	Length      int
	Prices      []PriceType
	Img         string
}

func GetPowerSpec(record LinkRecord) PowerSpec {
	powerData := PowerSpec{}

	// 处理中国区链接
	if strings.Contains(record.LinkCN, "zol") {
		record.LinkCN = getDetailsLinkFromZol(record.LinkCN, CreateCollector())
	}

	if record.LinkSpec != "" {
		powerData = getPowerSpecData(record.LinkSpec, CreateCollector())
	}

	powerData.Code = record.Name
	powerData.Brand = record.Brand
	if powerData.Name == "" {
		powerData.Name = record.Name
	}
	powerData.Name = RemoveBrandsFromName(powerData.Brand, powerData.Name)

	// 合并价格链接
	powerData.Prices = handleSpecPricesLogic(powerData.Prices, record)
	return powerData
}

func GetPowerData(spec PowerSpec) (PowerType, bool) {
	isValid := true
	newSpec := spec
	nameCN := spec.Name
	collector := CreateCollector()

	// 遍历所有价格数据进行处理
	for _, price := range newSpec.Prices {
		switch price.Region {
		case "CN":
			if strings.Contains(price.PriceLink, "zol") {
				tempSpec := getPowerSpecDataFromZol(price.PriceLink, collector)
				newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(PowerSpec)

				// 更新价格信息
				if updatedPrice := getPriceByPlatform(tempSpec.Prices, "CN", Platform_JD); updatedPrice != nil {
					isValid = isValid && checkPriceValid(updatedPrice.Price)
				}
			}
		case "US":
			tempSpec := getPowerUSPrice(price.PriceLink, collector)
			// 合并图片数据
			if newSpec.Img == "" && tempSpec.Img != "" {
				newSpec.Img = tempSpec.Img
			}
			// 合并规格数据
			newSpec = MergeStruct(newSpec, tempSpec, newSpec.Name).(PowerSpec)

			// 更新价格
			if updatedPrice := getPriceByPlatform(tempSpec.Prices, "US", Platform_Newegg); updatedPrice != nil {
				isValid = isValid && checkPriceValid(updatedPrice.Price)
			}
		}
	}

	return PowerType{
		Id:          SetProductId(spec.Brand, spec.Code),
		Brand:       newSpec.Brand,
		Name:        newSpec.Name,
		NameCN:      nameCN,
		ReleaseDate: newSpec.ReleaseDate,
		Standard:    newSpec.Standard,
		Wattage:     newSpec.Wattage,
		Size:        newSpec.Size,
		Modular:     newSpec.Modular,
		Efficiency:  newSpec.Efficiency,
		Length:      newSpec.Length,
		Prices:      deduplicatePrices(newSpec.Prices),
		Img:         newSpec.Img,
	}, isValid
}

func getPowerSpecData(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".content-wrapper", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".breadcrumb .active")
		specData.Img = element.ChildAttr(".tns-inner img", "src")
		specData.Prices = GetPriceLinkFromPangoly(element)

		element.ForEach(".table.table-striped tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("strong") {
			case "Release Date":
				specData.ReleaseDate = item.ChildText("td span")
			case "Wattage":
				specData.Wattage = extractNumberFromString(item.ChildTexts("td")[1])
			case "Type":
				specData.Size = item.ChildTexts("td")[1]
			case "Modular":
				specData.Modular = item.ChildTexts("td")[1]
			case "Efficiency":
				specData.Efficiency = item.ChildTexts("td")[1]
			case "Length":
				specData.Length = extractNumberFromString(item.ChildTexts("td")[1])
			}
		})
	})

	collector.Visit(link)

	return specData
}

func getPowerUSPrice(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		tempPrice := extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))
		available := element.ChildText(".row-side .product-buy-box .product-buy .btn-message")
		tempPrice = OutOfStockLogic(tempPrice, available)

		specData.Prices = upsertPrice(specData.Prices, PriceType{
			Region:    "US",
			Platform:  Platform_Newegg,
			Price:     tempPrice,
			PriceLink: link,
		})

		prdName := element.ChildText(".product-title")
		standard := extractATXStandard(prdName)

		element.ForEach(".tab-box .tab-panes tr", func(i int, item *colly.HTMLElement) {
			switch item.ChildText("th") {
			case "Type":
				standard = comparePSUStandard(standard, extractATXStandard(item.ChildText("td")))
			}
		})
		specData.Standard = standard
	})

	collector.Visit(link)
	return specData
}

func getPowerSpecDataFromZol(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{
		Standard: "ATX 2.0",
	}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		specData.Img = element.ChildAttr(".side .goods-card .goods-card__pic img", "src")
		specData.Prices = upsertPrice(specData.Prices, extractJDPriceFromZol(element))

		element.ForEach(".content table tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td span"))

			switch convertedHeader {
			case "电源版本":
				specData.Standard = extractATXStandard(convertedData)
			case "电源模组":
				if strings.Contains(convertedData, "全模组") {
					specData.Modular = "Full"
				} else if strings.Contains(convertedData, "半模组") {
					specData.Modular = "Semi"
				} else {
					specData.Modular = "No"
				}
			case "额定功率":
				specData.Wattage = extractNumberFromString(convertedData)
			case "80PLUS认证":
				if strings.Contains(convertedData, "钛金") {
					specData.Efficiency = "80+ Titanium"
				} else if strings.Contains(convertedData, "白金") || strings.Contains(convertedData, "铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(convertedData, "金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(convertedData, "银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(convertedData, "铜") {
					specData.Efficiency = "80+ Bronze"
				} else {
					specData.Efficiency = convertedData
				}
			}

			if strings.Contains(convertedHeader, "效率") {
				if specData.Efficiency == "" {
					specData.Efficiency = convertedData
				}
			}

			if strings.Contains(convertedHeader, "电源尺") {
				sizeStr := strings.Split(convertedData, "×")
				if strings.Contains(sizeStr[0], "125") {
					specData.Size = "SFX"
					sizeStr = removeElement(sizeStr, "125")
				} else {
					specData.Size = "ATX"
					sizeStr = removeElement(sizeStr, "150")
				}
				specData.Length = getTheLargestValueInArray(sizeStr)
			}
		})

	})
	collector.Visit(link)
	return specData
}

/*
func getPowerSpecDataFromHuntkey(link string, collector *colly.Collector) PowerSpec {
	specData := PowerSpec{
		Modular: "Full",
	}

	collectorErrorHandle(collector, link)
	collector.OnHTML(".border_hui", func(element *colly.HTMLElement) {
		specData.Name = element.ChildText(".lbProName")
		specData.Img = element.ChildAttr("#deta_pic .act img", "src")

		if strings.Contains(specData.Name, "JUMPER") || strings.Contains(specData.Name, "战鹰") || strings.Contains(specData.Name, "狼牙") {
			specData.Modular = "No"
		}
		if strings.HasPrefix(specData.Name, "ECO") || strings.HasPrefix(specData.Name, "GS") || strings.HasPrefix(specData.Name, "WD") || strings.HasPrefix(specData.Name, "GX") || strings.Contains(specData.Name, "猛擎") {
			specData.Modular = "No"
		}
		if strings.Contains(specData.Name, "MVP") {
			if specData.Name == "MVP500" || specData.Name == "MVP600" {
				specData.Modular = "Semi"
			} else {
				specData.Modular = "Full"
			}
		}
		if strings.HasPrefix(specData.Name, "HYPER") || strings.HasPrefix(specData.Name, "IP") || strings.Contains(specData.Name, "重火力") {
			specData.Modular = "Full"
		}
		if strings.Contains(specData.Name, "全模组") {
			specData.Modular = "Full"
		}

		fmt.Println(specData.Name)
		fmt.Println(specData.Modular)

		var mergedStrList []string

		element.ForEach("#con_one_2 tr td", func(i int, item *colly.HTMLElement) {
			if i%2 == 0 {
				mergedStrList = append(mergedStrList, item.Text)
			} else {
				setIndex := (i - 1) / 2
				mergedStrList[setIndex] = mergedStrList[setIndex] + item.Text
			}
		})

		for _, item := range mergedStrList {
			valStr := strings.Split(item, "：")[1]
			if strings.Contains(item, "额定功率") {
				specData.Wattage = extractNumberFromString(valStr)
			}
			if strings.Contains(item, "转换效率") {
				if strings.Contains(item, "80PLUS铜牌") {
					specData.Efficiency = "80+ Bronze"
				} else if strings.Contains(item, "80PLUS银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(item, "80PLUS金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(item, "80PLUS白金") || strings.Contains(item, "80PLUS铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(item, "80PLUS钛金") {
					specData.Efficiency = "80+ Titanium"
				} else {
					specData.Efficiency = valStr
				}
			}
			if strings.Contains(item, "尺寸") {
				sizeStr := strings.Split(item, "*")
				if strings.Contains(sizeStr[0], "125") {
					specData.Size = "SFX"
				} else {
					specData.Size = "ATX"
				}
				specData.Length = extractNumberFromString(sizeStr[2])
			}
		}
	})

	collector.Visit(link)
	return specData
}

func getPowerSpecDataFromPcOnline(link string, collector *colly.Collector) PowerSpec {
	detailsLink := strings.Split(link, ".htm")[0] + "_detail.html"
	specData := PowerSpec{}

	collectorErrorHandle(collector, detailsLink)
	collector.OnHTML(".area-detailparams", func(element *colly.HTMLElement) {
		element.ForEach(".bd-box tr", func(i int, item *colly.HTMLElement) {
			convertedHeader := convertGBKString(item.ChildText("th"))
			convertedData := convertGBKString(item.ChildText("td"))
			convertedLinkData := convertGBKString(item.ChildText("td .poptxt"))

			switch convertedHeader {
			case "型号":
				specData.Name = convertedData
			case "电源标准":
				specData.Size = convertedData
			case "额定功率":
				specData.Wattage = extractNumberFromString(convertedData)
			case "80PLUS认证":
				if strings.Contains(convertedLinkData, "铜牌") {
					specData.Efficiency = "80+ Bronze"
				} else if strings.Contains(convertedLinkData, "银牌") {
					specData.Efficiency = "80+ Silver"
				} else if strings.Contains(convertedLinkData, "金牌") {
					specData.Efficiency = "80+ Gold"
				} else if strings.Contains(convertedLinkData, "白金") || strings.Contains(convertedLinkData, "铂金") {
					specData.Efficiency = "80+ Platinum"
				} else if strings.Contains(convertedLinkData, "钛金") {
					specData.Efficiency = "80+ Titanium"
				} else {
					specData.Efficiency = convertedLinkData
				}
			}

			if strings.Contains(convertedHeader, "尺") {
				fmt.Println(convertedData)
				sizeStr := strings.Split(convertedData, "×")
				fmt.Println(sizeStr)
				specData.Length = extractNumberFromString(sizeStr[len(sizeStr)-1])
			}
		})

	})
	collector.Visit(detailsLink)
	return specData
}
*/

func extractATXStandard(name string) string {
	// 將名稱轉為大寫並去除空格
	normalized := strings.ToUpper(strings.ReplaceAll(name, " ", ""))

	// 定義正則表達式規則
	patterns := map[string]string{
		"ATX 3.1": `ATX(?:12V)?3\.1`,
		"ATX 3.0": `ATX(?:12V)?3\.0`,
		"ATX 3":   `ATX(?:12V)?3$`,
	}

	// 按順序匹配正則表達式
	for standard, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, normalized)
		if matched {
			return standard
		}
	}
	// 如果沒有匹配，返回默認的 ATX 2.0
	return "ATX 2.0"
}

func comparePSUStandard(str1 string, str2 string) string {
	if str1 == "ATX 2.0" {
		return str2
	}
	return str1
}
